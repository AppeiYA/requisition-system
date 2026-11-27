// Package s3 implements the attachment FileStorage port with an S3-compatible
// object store. Checksums travel with the object as metadata so Stat, Open and
// idempotent Move do not need a sidecar key.
package s3

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"

	attachmentkey "github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/storage/key"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

const (
	checksumMetadataKey = "attachment-checksum"
	defaultRegion       = "us-east-1"
)

var (
	_ attachmentout.FileStorage       = (*Storage)(nil)
	_ attachmentout.SignedURLProvider = (*Storage)(nil)

	errMissingBody = customerrors.NewBadRequestError("attachment body is required")
	errBadTTL      = customerrors.NewBadRequestError("signed URL TTL must be positive")
)

// Options is the driver-specific configuration the composition root passes to
// the adapter. It deliberately avoids importing the application config package.
type Options struct {
	Bucket          string
	Region          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	ForcePathStyle  bool
	UseSSL          bool
}

// Storage is an S3-backed FileStorage. The same instance also signs short-lived
// download URLs, so the HTTP layer never needs to know the bucket name.
type Storage struct {
	client    *awss3.Client
	presigner *awss3.PresignClient
	bucket    string
}

// New builds the AWS client. Static credentials are optional: when none are
// given the SDK's default chain (environment, shared config or task role) is
// used.
func New(opts Options) (*Storage, error) {
	bucket := strings.TrimSpace(opts.Bucket)
	if bucket == "" {
		return nil, fmt.Errorf("S3 bucket is required")
	}
	if (opts.AccessKeyID == "") != (opts.SecretAccessKey == "") {
		return nil, fmt.Errorf("S3 access key ID and secret access key must be set together")
	}

	region := strings.TrimSpace(opts.Region)
	if region == "" {
		region = defaultRegion
	}

	loadOptions := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(region)}
	if opts.AccessKeyID != "" {
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(opts.AccessKeyID, opts.SecretAccessKey, ""),
		))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	endpoint := normaliseEndpoint(opts)
	client := awss3.NewFromConfig(awsCfg, func(o *awss3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = opts.ForcePathStyle
	})

	return &Storage{
		client:    client,
		presigner: awss3.NewPresignClient(client),
		bucket:    bucket,
	}, nil
}

func normaliseEndpoint(opts Options) string {
	endpoint := strings.TrimSpace(opts.Endpoint)
	if endpoint == "" {
		return ""
	}
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}
	scheme := "https"
	if !opts.UseSSL {
		scheme = "http"
	}
	return scheme + "://" + endpoint
}

// Put stores the object and its checksum metadata. The reader is fully consumed
// before the upload starts, so a read failure never leaves a partial object.
func (s *Storage) Put(
	ctx context.Context,
	key string,
	r io.Reader,
	opts attachmentout.PutOptions,
) (attachmentout.ObjectInfo, error) {
	if err := attachmentkey.ValidateKey(key); err != nil {
		return attachmentout.ObjectInfo{}, err
	}
	if r == nil {
		return attachmentout.ObjectInfo{}, errMissingBody
	}
	if err := ctx.Err(); err != nil {
		return attachmentout.ObjectInfo{}, err
	}

	var payload bytes.Buffer
	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(&payload, hasher), &contextReader{ctx: ctx, r: r})
	if err != nil {
		return attachmentout.ObjectInfo{}, storageFailure("read object", err)
	}
	if err := ctx.Err(); err != nil {
		return attachmentout.ObjectInfo{}, err
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	_, err = s.client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(payload.Bytes()),
		ContentType: aws.String(strings.TrimSpace(opts.ContentType)),
		Metadata:    map[string]string{checksumMetadataKey: checksum},
	})
	if err != nil {
		return attachmentout.ObjectInfo{}, storageFailure("store object", err)
	}

	return attachmentout.ObjectInfo{
		Key:         key,
		Size:        written,
		ContentType: strings.TrimSpace(opts.ContentType),
		Checksum:    checksum,
		ModifiedAt:  time.Now().UTC(),
	}, nil
}

// Open returns a seekable in-memory reader. S3 bodies are not seekable, and
// http.ServeContent needs seeking for byte ranges; attachment bodies are capped
// well below memory limits by the upload use cases.
func (s *Storage) Open(ctx context.Context, key string) (io.ReadSeekCloser, attachmentout.ObjectInfo, error) {
	if err := attachmentkey.ValidateKey(key); err != nil {
		return nil, attachmentout.ObjectInfo{}, err
	}
	if err := ctx.Err(); err != nil {
		return nil, attachmentout.ObjectInfo{}, err
	}

	out, err := s.client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, attachmentout.ObjectInfo{}, mapStorageError(err)
	}
	defer out.Body.Close()

	payload, err := io.ReadAll(&contextReader{ctx: ctx, r: out.Body})
	if err != nil {
		return nil, attachmentout.ObjectInfo{}, storageFailure("read object", err)
	}
	checksum := metadataChecksum(out.Metadata)
	if checksum == "" {
		checksum = sha256Hex(payload)
	}
	info := attachmentout.ObjectInfo{
		Key:         key,
		Size:        int64(len(payload)),
		ContentType: aws.ToString(out.ContentType),
		Checksum:    checksum,
		ModifiedAt:  aws.ToTime(out.LastModified),
	}
	if info.ModifiedAt.IsZero() {
		info.ModifiedAt = time.Now().UTC()
	}
	return &memoryReadSeekCloser{Reader: bytes.NewReader(payload)}, info, nil
}

// Stat reads object metadata with a HEAD request. Objects written outside this
// adapter have no checksum metadata, so their bytes are read once to compute it.
func (s *Storage) Stat(ctx context.Context, key string) (attachmentout.ObjectInfo, error) {
	if err := attachmentkey.ValidateKey(key); err != nil {
		return attachmentout.ObjectInfo{}, err
	}

	info, err := s.head(ctx, key)
	if err != nil {
		return attachmentout.ObjectInfo{}, err
	}
	if info.Checksum == "" {
		verified, err := s.verifiedInfo(ctx, key, "")
		if err != nil {
			return attachmentout.ObjectInfo{}, err
		}
		info.Checksum = verified.Checksum
	}
	return info, nil
}

// Delete removes an object. S3 treats a missing key as a successful delete.
func (s *Storage) Delete(ctx context.Context, key string) error {
	if err := attachmentkey.ValidateKey(key); err != nil {
		return err
	}
	if _, err := s.client.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}); err != nil {
		if attachmentout.IsObjectNotFound(mapStorageError(err)) {
			return nil
		}
		return storageFailure("delete object", err)
	}
	return nil
}

// Move is idempotent. A retry with a missing source and an existing destination
// verifies the destination checksum instead of failing. If both objects exist,
// the destination is authoritative and the staging leftover is removed.
func (s *Storage) Move(
	ctx context.Context,
	from, to, expectedChecksum string,
) (attachmentout.ObjectInfo, error) {
	if err := attachmentkey.ValidateKey(from); err != nil {
		return attachmentout.ObjectInfo{}, err
	}
	if err := attachmentkey.ValidateKey(to); err != nil {
		return attachmentout.ObjectInfo{}, err
	}
	if from == to {
		return attachmentout.ObjectInfo{}, attachmentkey.ErrInvalidKey
	}

	sourceInfo, sourceErr := s.head(ctx, from)
	sourceMissing := attachmentout.IsObjectNotFound(sourceErr)
	if sourceErr != nil && !sourceMissing {
		return attachmentout.ObjectInfo{}, sourceErr
	}
	_, destErr := s.head(ctx, to)
	destMissing := attachmentout.IsObjectNotFound(destErr)
	if destErr != nil && !destMissing {
		return attachmentout.ObjectInfo{}, destErr
	}

	switch {
	case sourceMissing && destMissing:
		return attachmentout.ObjectInfo{}, attachmentout.ErrObjectNotFound
	case sourceMissing:
		return s.verifiedInfo(ctx, to, expectedChecksum)
	case destMissing:
		return s.moveObject(ctx, from, to, expectedChecksum, sourceInfo.ContentType)
	default:
		info, err := s.verifiedInfo(ctx, to, expectedChecksum)
		if err != nil {
			return attachmentout.ObjectInfo{}, err
		}
		_ = s.Delete(ctx, from)
		return info, nil
	}
}

func (s *Storage) head(ctx context.Context, key string) (attachmentout.ObjectInfo, error) {
	out, err := s.client.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return attachmentout.ObjectInfo{}, mapStorageError(err)
	}
	info := attachmentout.ObjectInfo{
		Key:         key,
		Size:        aws.ToInt64(out.ContentLength),
		ContentType: aws.ToString(out.ContentType),
		Checksum:    metadataChecksum(out.Metadata),
		ModifiedAt:  aws.ToTime(out.LastModified),
	}
	if info.ModifiedAt.IsZero() {
		info.ModifiedAt = time.Now().UTC()
	}
	return info, nil
}

func (s *Storage) moveObject(
	ctx context.Context,
	from, to, expectedChecksum, fallbackContentType string,
) (attachmentout.ObjectInfo, error) {
	out, err := s.client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(from),
	})
	if err != nil {
		return attachmentout.ObjectInfo{}, mapStorageError(err)
	}
	defer out.Body.Close()

	payload, err := io.ReadAll(&contextReader{ctx: ctx, r: out.Body})
	if err != nil {
		return attachmentout.ObjectInfo{}, storageFailure("read source object", err)
	}
	checksum := sha256Hex(payload)
	if expectedChecksum != "" && checksum != expectedChecksum {
		return attachmentout.ObjectInfo{}, fmt.Errorf("move object: %w", attachmentout.ErrChecksumMismatch)
	}
	contentType := aws.ToString(out.ContentType)
	if contentType == "" {
		contentType = fallbackContentType
	}

	_, err = s.client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(to),
		Body:        bytes.NewReader(payload),
		ContentType: aws.String(contentType),
		Metadata:    map[string]string{checksumMetadataKey: checksum},
	})
	if err != nil {
		return attachmentout.ObjectInfo{}, storageFailure("store moved object", err)
	}
	if err := s.Delete(ctx, from); err != nil {
		return attachmentout.ObjectInfo{}, err
	}

	return attachmentout.ObjectInfo{
		Key:         to,
		Size:        int64(len(payload)),
		ContentType: contentType,
		Checksum:    checksum,
		ModifiedAt:  time.Now().UTC(),
	}, nil
}

// verifiedInfo trusts the bytes rather than metadata when completing a retry.
func (s *Storage) verifiedInfo(
	ctx context.Context,
	key, expectedChecksum string,
) (attachmentout.ObjectInfo, error) {
	out, err := s.client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return attachmentout.ObjectInfo{}, mapStorageError(err)
	}
	defer out.Body.Close()

	payload, err := io.ReadAll(&contextReader{ctx: ctx, r: out.Body})
	if err != nil {
		return attachmentout.ObjectInfo{}, storageFailure("read object", err)
	}
	checksum := sha256Hex(payload)
	if expectedChecksum != "" && checksum != expectedChecksum {
		return attachmentout.ObjectInfo{}, fmt.Errorf("move object: %w", attachmentout.ErrChecksumMismatch)
	}
	info := attachmentout.ObjectInfo{
		Key:         key,
		Size:        int64(len(payload)),
		ContentType: aws.ToString(out.ContentType),
		Checksum:    checksum,
		ModifiedAt:  aws.ToTime(out.LastModified),
	}
	if info.ModifiedAt.IsZero() {
		info.ModifiedAt = time.Now().UTC()
	}
	return info, nil
}

// SignedURL hands the HTTP layer a presigned GET URL. The existence check keeps
// a deleted object from producing a URL that would only fail for the client.
func (s *Storage) SignedURL(
	ctx context.Context,
	key string,
	ttl time.Duration,
	downloadName string,
) (string, time.Time, error) {
	if err := attachmentkey.ValidateKey(key); err != nil {
		return "", time.Time{}, err
	}
	if ttl <= 0 {
		return "", time.Time{}, errBadTTL
	}
	if _, err := s.head(ctx, key); err != nil {
		return "", time.Time{}, err
	}

	request, err := s.presigner.PresignGetObject(ctx, &awss3.GetObjectInput{
		Bucket:                     aws.String(s.bucket),
		Key:                        aws.String(key),
		ResponseContentDisposition: aws.String(contentDisposition(downloadName)),
	}, awss3.WithPresignExpires(ttl))
	if err != nil {
		return "", time.Time{}, storageFailure("sign download URL", err)
	}
	return request.URL, time.Now().UTC().Add(ttl), nil
}

func contentDisposition(name string) string {
	name = strings.TrimSpace(name)
	name = strings.NewReplacer(`\`, "_", "/", "_", `"`, "_").Replace(name)
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return '_'
		}
		return r
	}, name)
	if name == "" {
		name = "attachment"
	}
	return `attachment; filename="` + name + `"`
}

func metadataChecksum(metadata map[string]string) string {
	for key, value := range metadata {
		if strings.EqualFold(key, checksumMetadataKey) ||
			strings.EqualFold(key, "x-amz-meta-"+checksumMetadataKey) {
			return strings.ToLower(strings.TrimSpace(value))
		}
	}
	return ""
}

func sha256Hex(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func storageFailure(action string, err error) error {
	return fmt.Errorf("%s: %w", action, mapStorageError(err))
}

func mapStorageError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchKey", "NotFound", "404":
			return attachmentout.ErrObjectNotFound
		}
	}
	var responseErr *smithyhttp.ResponseError
	if errors.As(err, &responseErr) && responseErr.HTTPStatusCode() == http.StatusNotFound {
		return attachmentout.ErrObjectNotFound
	}
	return fmt.Errorf("%w: %v", attachmentout.ErrStorageUnavailable, err)
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

type memoryReadSeekCloser struct {
	*bytes.Reader
}

func (*memoryReadSeekCloser) Close() error { return nil }

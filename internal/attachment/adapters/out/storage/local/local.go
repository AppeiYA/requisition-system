// Package local implements the attachment FileStorage port with files on disk.
package local

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	attachmentkey "github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/storage/key"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

const (
	dirMode  os.FileMode = 0o700
	fileMode os.FileMode = 0o600

	metaSuffix    = ".meta"
	tmpSuffix     = ".tmp"
	metaTmpSuffix = ".meta.tmp"
)

var (
	_ attachmentout.FileStorage = (*Storage)(nil)

	errMissingBody = customerrors.NewBadRequestError("attachment body is required")
)

// Storage stores objects below a single root directory. Object keys are
// validated before they are ever joined to the root, so no client input can
// escape that directory.
type Storage struct {
	root    string
	initErr error
}

// New creates (if needed) the storage root with 0700 permissions.
func New(root string) *Storage {
	cleaned := filepath.Clean(root)
	s := &Storage{root: cleaned}
	if root == "" {
		s.initErr = storageFailure("initialise storage root")
		return s
	}
	if err := os.MkdirAll(cleaned, dirMode); err != nil {
		s.initErr = storageFailure("initialise storage root")
		return s
	}
	if err := os.Chmod(cleaned, dirMode); err != nil {
		s.initErr = storageFailure("initialise storage root")
		return s
	}
	return s
}

// objectMeta is the small sidecar written next to every object. It keeps Stat
// from having to read the object bytes just to report its metadata.
type objectMeta struct {
	ContentType string `json:"content_type"`
	Checksum    string `json:"checksum"`
	Size        int64  `json:"size"`
}

// Put writes the object to a temporary file and renames it into place only
// after every byte has been copied successfully. Any error leaves no object,
// sidecar or temporary file behind.
func (s *Storage) Put(ctx context.Context, key string, r io.Reader, opts attachmentout.PutOptions) (attachmentout.ObjectInfo, error) {
	var info attachmentout.ObjectInfo
	if err := s.ready(ctx); err != nil {
		return info, err
	}
	if err := attachmentkey.ValidateKey(key); err != nil {
		return info, err
	}
	if r == nil {
		return info, errMissingBody
	}

	objectPath := s.objectPath(key)
	metaPath := objectPath + metaSuffix
	objectTmp := objectPath + tmpSuffix
	metaTmp := objectPath + metaTmpSuffix

	dir := filepath.Dir(objectPath)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return info, storageFailure("prepare storage directory")
	}
	if err := os.Chmod(dir, dirMode); err != nil {
		return info, storageFailure("prepare storage directory")
	}

	f, err := os.OpenFile(objectTmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return info, storageFailure("create temporary object")
	}
	if err := f.Chmod(fileMode); err != nil {
		_ = f.Close()
		_ = os.Remove(objectTmp)
		return info, storageFailure("set object permissions")
	}

	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(f, hasher), &contextReader{ctx: ctx, r: r})
	if err != nil {
		_ = f.Close()
		_ = os.Remove(objectTmp)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return info, ctxErr
		}
		return info, storageFailure("write object")
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(objectTmp)
		return info, storageFailure("flush object")
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(objectTmp)
		return info, storageFailure("close object")
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	meta := objectMeta{
		ContentType: opts.ContentType,
		Checksum:    checksum,
		Size:        written,
	}

	if err := writeMeta(ctx, metaTmp, meta); err != nil {
		_ = os.Remove(objectTmp)
		return info, err
	}

	objectRenamed, metaRenamed, committed := false, false, false
	defer func() {
		if !objectRenamed {
			_ = os.Remove(objectTmp)
		} else if !committed {
			_ = os.Remove(objectPath)
		}
		if !metaRenamed {
			_ = os.Remove(metaTmp)
		} else if !committed {
			_ = os.Remove(metaPath)
		}
	}()

	if err := ctx.Err(); err != nil {
		return info, err
	}
	if err := os.Rename(objectTmp, objectPath); err != nil {
		return info, storageFailure("store object")
	}
	objectRenamed = true

	if err := ctx.Err(); err != nil {
		return info, err
	}
	if err := os.Rename(metaTmp, metaPath); err != nil {
		return info, storageFailure("store object metadata")
	}
	metaRenamed = true
	_ = syncDir(dir)

	if err := ctx.Err(); err != nil {
		return info, err
	}
	st, err := os.Stat(objectPath)
	if err != nil {
		return info, storageFailure("stat stored object")
	}

	committed = true
	return attachmentout.ObjectInfo{
		Key:         key,
		Size:        written,
		ContentType: opts.ContentType,
		Checksum:    checksum,
		ModifiedAt:  st.ModTime(),
	}, nil
}

// Open returns a seekable *os.File so the HTTP layer can serve byte ranges.
func (s *Storage) Open(ctx context.Context, key string) (io.ReadSeekCloser, attachmentout.ObjectInfo, error) {
	var info attachmentout.ObjectInfo
	if err := s.ready(ctx); err != nil {
		return nil, info, err
	}
	if err := attachmentkey.ValidateKey(key); err != nil {
		return nil, info, err
	}

	f, err := os.Open(s.objectPath(key))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, info, notFound()
		}
		return nil, info, storageFailure("open object")
	}

	info, err = s.Stat(ctx, key)
	if err != nil {
		_ = f.Close()
		return nil, attachmentout.ObjectInfo{}, err
	}
	return f, info, nil
}

// Stat reports object metadata without reading the object whenever the
// sidecar is present. A sidecar-less object still gets a computed checksum.
func (s *Storage) Stat(ctx context.Context, key string) (attachmentout.ObjectInfo, error) {
	var info attachmentout.ObjectInfo
	if err := s.ready(ctx); err != nil {
		return info, err
	}
	if err := attachmentkey.ValidateKey(key); err != nil {
		return info, err
	}

	objectPath := s.objectPath(key)
	st, err := os.Stat(objectPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return info, notFound()
		}
		return info, storageFailure("stat object")
	}

	info = attachmentout.ObjectInfo{
		Key:        key,
		Size:       st.Size(),
		ModifiedAt: st.ModTime(),
	}

	data, err := os.ReadFile(objectPath + metaSuffix)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return info, storageFailure("read object metadata")
		}
		checksum, err := checksumFile(ctx, objectPath)
		if err != nil {
			return info, err
		}
		info.Checksum = checksum
		return info, nil
	}

	var meta objectMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return info, storageFailure("decode object metadata")
	}
	info.ContentType = meta.ContentType
	info.Checksum = meta.Checksum
	if info.Checksum == "" {
		checksum, err := checksumFile(ctx, objectPath)
		if err != nil {
			return info, err
		}
		info.Checksum = checksum
	}
	return info, nil
}

// Delete removes the object and all of its companion files. A missing object
// is not an error.
func (s *Storage) Delete(ctx context.Context, key string) error {
	if err := s.ready(ctx); err != nil {
		return err
	}
	if err := attachmentkey.ValidateKey(key); err != nil {
		return err
	}

	objectPath := s.objectPath(key)
	for _, path := range []string{
		objectPath,
		objectPath + metaSuffix,
		objectPath + tmpSuffix,
		objectPath + metaTmpSuffix,
	} {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return storageFailure("delete object")
		}
	}
	return nil
}

// Move relocates a staged object to its permanent key. The object bytes are
// renamed first, then a fresh metadata sidecar is written from the bytes that
// were actually moved. A retry after a crash in either step sees the object
// already at the destination and verifies its checksum instead of failing.
func (s *Storage) Move(ctx context.Context, from, to, expectedChecksum string) (attachmentout.ObjectInfo, error) {
	var info attachmentout.ObjectInfo
	if err := s.ready(ctx); err != nil {
		return info, err
	}
	if err := attachmentkey.ValidateKey(from); err != nil {
		return info, err
	}
	if err := attachmentkey.ValidateKey(to); err != nil {
		return info, err
	}
	if from == to {
		return info, attachmentkey.ErrInvalidKey
	}

	fromPath := s.objectPath(from)
	toPath := s.objectPath(to)
	dir := filepath.Dir(toPath)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return info, storageFailure("prepare storage directory")
	}
	if err := os.Chmod(dir, dirMode); err != nil {
		return info, storageFailure("prepare storage directory")
	}

	_, sourceErr := os.Stat(fromPath)
	sourceMissing := errors.Is(sourceErr, os.ErrNotExist)
	if sourceErr != nil && !sourceMissing {
		return info, storageFailure("stat source object")
	}
	_, destErr := os.Stat(toPath)
	destMissing := errors.Is(destErr, os.ErrNotExist)
	if destErr != nil && !destMissing {
		return info, storageFailure("stat destination object")
	}

	switch {
	case sourceMissing && destMissing:
		return info, notFound()
	case sourceMissing:
		info, err := s.verifiedInfo(ctx, to, expectedChecksum)
		if err != nil {
			return attachmentout.ObjectInfo{}, err
		}
		return info, nil
	case destMissing:
		return s.moveObject(ctx, from, to, fromPath, toPath, expectedChecksum)
	default:
		// Another worker already moved the object. Verify it, then remove any
		// staging leftovers so cleanup has nothing to collect.
		info, err := s.verifiedInfo(ctx, to, expectedChecksum)
		if err != nil {
			return attachmentout.ObjectInfo{}, err
		}
		s.removeObjectFiles(fromPath)
		return info, nil
	}
}

func (s *Storage) moveObject(
	ctx context.Context,
	from, to, fromPath, toPath, expectedChecksum string,
) (attachmentout.ObjectInfo, error) {
	var info attachmentout.ObjectInfo
	sourceInfo, err := s.Stat(ctx, from)
	if err != nil {
		return info, err
	}
	checksum, err := checksumFile(ctx, fromPath)
	if err != nil {
		return info, err
	}
	if expectedChecksum != "" && checksum != expectedChecksum {
		return info, fmt.Errorf("move object: %w", attachmentout.ErrChecksumMismatch)
	}

	metaPath := toPath + metaSuffix
	metaTmpPath := toPath + metaTmpSuffix
	objectMoved, metaMoved, committed := false, false, false
	defer func() {
		if !committed {
			if metaMoved {
				_ = os.Remove(metaPath)
			}
			_ = os.Remove(metaTmpPath)
			if objectMoved {
				_ = os.Rename(toPath, fromPath)
			}
		}
	}()

	if err := ctx.Err(); err != nil {
		return info, err
	}
	if err := os.Rename(fromPath, toPath); err != nil {
		return info, storageFailure("move object")
	}
	objectMoved = true

	meta := objectMeta{
		ContentType: sourceInfo.ContentType,
		Checksum:    checksum,
		Size:        sourceInfo.Size,
	}
	if err := writeMeta(ctx, metaTmpPath, meta); err != nil {
		return info, err
	}
	if err := os.Rename(metaTmpPath, metaPath); err != nil {
		return info, storageFailure("store moved object metadata")
	}
	metaMoved = true
	_ = syncDir(filepath.Dir(toPath))

	s.removeObjectFiles(fromPath)
	if err := ctx.Err(); err != nil {
		return info, err
	}
	st, err := os.Stat(toPath)
	if err != nil {
		return info, storageFailure("stat moved object")
	}
	committed = true
	return attachmentout.ObjectInfo{
		Key:         to,
		Size:        st.Size(),
		ContentType: sourceInfo.ContentType,
		Checksum:    checksum,
		ModifiedAt:  st.ModTime(),
	}, nil
}

// verifiedInfo trusts the bytes, not the sidecar, when completing an
// interrupted move. This is what makes a crash between rename and metadata
// write recoverable.
func (s *Storage) verifiedInfo(
	ctx context.Context,
	key, expectedChecksum string,
) (attachmentout.ObjectInfo, error) {
	info, err := s.Stat(ctx, key)
	if err != nil {
		return attachmentout.ObjectInfo{}, err
	}
	checksum, err := checksumFile(ctx, s.objectPath(key))
	if err != nil {
		return attachmentout.ObjectInfo{}, err
	}
	if expectedChecksum != "" && checksum != expectedChecksum {
		return attachmentout.ObjectInfo{}, fmt.Errorf("move object: %w", attachmentout.ErrChecksumMismatch)
	}
	info.Checksum = checksum
	return info, nil
}

// removeObjectFiles clears the source sidecar and any interrupted temporary
// files after the object itself has been moved.
func (s *Storage) removeObjectFiles(objectPath string) {
	for _, path := range []string{
		objectPath + metaSuffix,
		objectPath + tmpSuffix,
		objectPath + metaTmpSuffix,
	} {
		_ = os.Remove(path)
	}
}

func (s *Storage) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.initErr
}

func (s *Storage) objectPath(key string) string {
	return filepath.Join(s.root, filepath.FromSlash(key))
}

func writeMeta(ctx context.Context, path string, meta objectMeta) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		_ = os.Remove(path)
		return storageFailure("create object metadata")
	}
	if err := f.Chmod(fileMode); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return storageFailure("set object metadata permissions")
	}
	err = json.NewEncoder(f).Encode(meta)
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(path)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return storageFailure("write object metadata")
	}
	return nil
}

func checksumFile(ctx context.Context, path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", notFound()
		}
		return "", storageFailure("open object for checksum")
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, &contextReader{ctx: ctx, r: f}); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", ctxErr
		}
		return "", storageFailure("checksum object")
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// contextReader stops a long copy as soon as the caller cancels.
type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (c *contextReader) Read(p []byte) (int, error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}
	return c.r.Read(p)
}

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = d.Sync()
	if closeErr := d.Close(); err == nil {
		err = closeErr
	}
	return err
}

func storageFailure(op string) error {
	// Keep filesystem paths out of errors: handlers may surface the message.
	return fmt.Errorf("%s: %w", op, attachmentout.ErrStorageUnavailable)
}

func notFound() error {
	return fmt.Errorf("object lookup: %w", attachmentout.ErrObjectNotFound)
}

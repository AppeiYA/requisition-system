package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentmocks "github.com/AppeiYA/requisition-system/internal/attachment/mocks"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

func attachmentRecordWithSize(size int64) *domain.Attachment {
	ref, _ := domain.NewAttachmentRef(domain.RefNameRequisition, attachmentmocks.DefaultRefID)
	name, _ := domain.NewAttachmentFileName(attachmentmocks.DefaultFileName)
	attachment, _ := domain.NewAttachment(
		attachmentmocks.DefaultAttachmentID,
		ref,
		attachmentmocks.DefaultStorageKey,
		name,
		domain.ContentTypePDF,
		size,
		strings.Repeat("b", domain.ChecksumLength),
		attachmentmocks.DefaultActorID,
	)
	return attachment
}

const downloadTestTTL = 3 * time.Minute

func downloadTestUsecase(
	repo *attachmentmocks.MockAttachmentRepository,
	storage *attachmentmocks.MockFileStorage,
	access *attachmentmocks.MockRefAccessChecker,
	signer attachmentout.SignedURLProvider,
) *DownloadAttachmentUsecase {
	return NewDownloadAttachmentUsecase(repo, storage, access, signer, downloadTestTTL)
}

func downloadTestQuery() attachmentin.DownloadAttachmentQuery {
	return attachmentin.DownloadAttachmentQuery{
		AttachmentID: attachmentmocks.DefaultAttachmentID,
		ActorID:      attachmentmocks.DefaultActorID,
		ActorRole:    attachmentmocks.DefaultActorRole,
	}
}

type downloadTestCloser struct {
	io.ReadSeeker
	closed bool
}

func (c *downloadTestCloser) Close() error {
	c.closed = true
	return nil
}

type nopReadSeekCloser struct {
	*bytes.Reader
}

func (nopReadSeekCloser) Close() error { return nil }

func TestDownloadAttachmentUsecaseStreamsWhenNoSigner(t *testing.T) {
	repo := attachmentmocks.NewMockAttachmentRepository()
	storage := attachmentmocks.NewMockFileStorage()
	access := attachmentmocks.NewMockRefAccessChecker()
	payload := []byte("hello attachment")

	var openedKey string
	var accessAction attachmentout.Action
	repo.GetByIDFunc = func(_ context.Context, id string) (*domain.Attachment, error) {
		if id != attachmentmocks.DefaultAttachmentID {
			t.Errorf("GetByID(%q)", id)
		}
		return attachmentRecordWithSize(int64(len(payload))), nil
	}
	access.CanAccessFunc = func(_ context.Context, refName, refID, actorID, actorRole string, action attachmentout.Action) error {
		accessAction = action
		if refName != attachmentmocks.DefaultRefName || refID != attachmentmocks.DefaultRefID ||
			actorID != attachmentmocks.DefaultActorID || actorRole != attachmentmocks.DefaultActorRole {
			t.Errorf("unexpected access arguments: %s/%s/%s/%s", refName, refID, actorID, actorRole)
		}
		return nil
	}
	storage.OpenFunc = func(_ context.Context, key string) (io.ReadSeekCloser, attachmentout.ObjectInfo, error) {
		openedKey = key
		return nopReadSeekCloser{bytes.NewReader(payload)}, attachmentout.ObjectInfo{Key: key, Size: int64(len(payload))}, nil
	}

	download, err := downloadTestUsecase(repo, storage, access, nil).Execute(context.Background(), downloadTestQuery())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if download.Attachment == nil || download.Attachment.ID() != attachmentmocks.DefaultAttachmentID {
		t.Fatalf("unexpected attachment: %v", download.Attachment)
	}
	if download.Stream == nil {
		t.Fatal("expected a stream when no signer is configured")
	}
	defer download.Stream.Close()
	if download.Location != "" || !download.ExpiresAt.IsZero() {
		t.Errorf("unexpected redirect fields: %q/%v", download.Location, download.ExpiresAt)
	}
	if openedKey != attachmentmocks.DefaultStorageKey {
		t.Errorf("Open key = %q", openedKey)
	}
	if accessAction != attachmentout.ActionRead {
		t.Errorf("access action = %q, want read", accessAction)
	}
	got, err := io.ReadAll(download.Stream)
	if err != nil {
		t.Fatalf("reading stream: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("stream = %q, want %q", got, payload)
	}
}

func TestDownloadAttachmentUsecaseRedirectsWithSigner(t *testing.T) {
	repo := attachmentmocks.NewMockAttachmentRepository()
	storage := attachmentmocks.NewMockFileStorage()
	access := attachmentmocks.NewMockRefAccessChecker()
	signer := attachmentmocks.NewMockSignedURLProvider()
	expires := time.Now().Add(downloadTestTTL)

	var (
		signedKey, signedName string
		signedTTL             time.Duration
		openCalled            bool
	)
	signer.SignedURLFunc = func(_ context.Context, key string, ttl time.Duration, downloadName string) (string, time.Time, error) {
		signedKey, signedTTL, signedName = key, ttl, downloadName
		return "https://storage.example.test/signed.pdf", expires, nil
	}
	storage.OpenFunc = func(context.Context, string) (io.ReadSeekCloser, attachmentout.ObjectInfo, error) {
		openCalled = true
		return nil, attachmentout.ObjectInfo{}, nil
	}

	download, err := downloadTestUsecase(repo, storage, access, signer).Execute(context.Background(), downloadTestQuery())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if download.Stream != nil {
		t.Error("stream must be nil for a signed URL")
	}
	if download.Location != "https://storage.example.test/signed.pdf" || !download.ExpiresAt.Equal(expires) {
		t.Errorf("download = %q/%v", download.Location, download.ExpiresAt)
	}
	if signedKey != attachmentmocks.DefaultStorageKey || signedTTL != downloadTestTTL || signedName != attachmentmocks.DefaultFileName {
		t.Errorf("SignedURL(%q, %v, %q)", signedKey, signedTTL, signedName)
	}
	if openCalled {
		t.Error("storage was opened despite a signer being configured")
	}
}

func TestDownloadAttachmentUsecaseAccessDenied(t *testing.T) {
	repo := attachmentmocks.NewMockAttachmentRepository()
	storage := attachmentmocks.NewMockFileStorage()
	access := attachmentmocks.NewMockRefAccessChecker()
	denied := customerrors.NewForbiddenError("no read access")
	access.CanAccessFunc = func(context.Context, string, string, string, string, attachmentout.Action) error {
		return denied
	}
	openCalled := false
	storage.OpenFunc = func(context.Context, string) (io.ReadSeekCloser, attachmentout.ObjectInfo, error) {
		openCalled = true
		return nil, attachmentout.ObjectInfo{}, nil
	}

	_, err := downloadTestUsecase(repo, storage, access, nil).Execute(context.Background(), downloadTestQuery())
	if !errors.Is(err, denied) {
		t.Fatalf("error = %v, want access error", err)
	}
	if openCalled {
		t.Error("bytes were opened after access was denied")
	}
}

func TestDownloadAttachmentUsecaseMappingAndIntegrity(t *testing.T) {
	t.Run("missing metadata row", func(t *testing.T) {
		repo := attachmentmocks.NewMockAttachmentRepository()
		repo.GetByIDFunc = func(context.Context, string) (*domain.Attachment, error) {
			return nil, customerrors.NewNotFoundError("row missing")
		}
		_, err := downloadTestUsecase(repo, attachmentmocks.NewMockFileStorage(), attachmentmocks.NewMockRefAccessChecker(), nil).
			Execute(context.Background(), downloadTestQuery())
		if err != ErrAttachmentNotFound {
			t.Fatalf("error = %v, want ErrAttachmentNotFound", err)
		}
	})

	t.Run("missing object", func(t *testing.T) {
		storage := attachmentmocks.NewMockFileStorage()
		storage.OpenFunc = func(context.Context, string) (io.ReadSeekCloser, attachmentout.ObjectInfo, error) {
			return nil, attachmentout.ObjectInfo{}, attachmentout.ErrObjectNotFound
		}
		_, err := downloadTestUsecase(attachmentmocks.NewMockAttachmentRepository(), storage, attachmentmocks.NewMockRefAccessChecker(), nil).
			Execute(context.Background(), downloadTestQuery())
		if !errors.Is(err, attachmentout.ErrObjectNotFound) {
			t.Fatalf("error = %v, want object-not-found", err)
		}
		if ce, ok := err.(*customerrors.CustomError); !ok || ce.Code != 404 {
			t.Fatalf("error code = %v, want 404", err)
		}
	})

	t.Run("size mismatch is refused", func(t *testing.T) {
		storage := attachmentmocks.NewMockFileStorage()
		stream := &downloadTestCloser{ReadSeeker: strings.NewReader("too long")}
		storage.OpenFunc = func(context.Context, string) (io.ReadSeekCloser, attachmentout.ObjectInfo, error) {
			return stream, attachmentout.ObjectInfo{Size: attachmentmocks.DefaultFileSize + 1}, nil
		}
		_, err := downloadTestUsecase(attachmentmocks.NewMockAttachmentRepository(), storage, attachmentmocks.NewMockRefAccessChecker(), nil).
			Execute(context.Background(), downloadTestQuery())
		if err != ErrAttachmentIntegrity {
			t.Fatalf("error = %v, want ErrAttachmentIntegrity", err)
		}
		if !stream.closed {
			t.Error("mismatched stream was not closed")
		}
	})
}

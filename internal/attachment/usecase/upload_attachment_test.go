package usecase

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentmocks "github.com/AppeiYA/requisition-system/internal/attachment/mocks"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

const (
	uploadTestMaxSize   = int64(64)
	uploadTestMaxPerRef = 3
)

var uploadTestBody = []byte("%PDF-1.4\nupload body\n")

type uploadHarness struct {
	repo     *attachmentmocks.MockAttachmentRepository
	storage  *attachmentmocks.MockFileStorage
	detector *attachmentmocks.MockContentTypeDetector
	access   *attachmentmocks.MockRefAccessChecker
	idGen    *attachmentmocks.MockIDGenerator
	tx       *attachmentmocks.MockTransactionManager
	uc       *UploadAttachmentUsecase
}

func newUploadHarness() *uploadHarness {
	h := &uploadHarness{
		repo:     attachmentmocks.NewMockAttachmentRepository(),
		storage:  attachmentmocks.NewMockFileStorage(),
		detector: attachmentmocks.NewMockContentTypeDetector(),
		access:   attachmentmocks.NewMockRefAccessChecker(),
		idGen:    attachmentmocks.NewMockIDGenerator(),
		tx:       attachmentmocks.NewMockTransactionManager(),
	}
	h.uc = NewUploadAttachmentUsecase(
		h.repo, h.storage, h.detector, h.access, h.idGen, h.tx,
		uploadTestMaxSize, uploadTestMaxPerRef,
	)
	return h
}

func validUploadCommand(body []byte) attachmentin.UploadAttachmentCommand {
	return attachmentin.UploadAttachmentCommand{
		ActorID:   attachmentmocks.DefaultActorID,
		ActorRole: attachmentmocks.DefaultActorRole,
		RefName:   attachmentmocks.DefaultRefName,
		RefID:     attachmentmocks.DefaultRefID,
		FileName:  attachmentmocks.DefaultFileName,
		Body:      bytes.NewReader(body),
		Size:      int64(len(body)),
	}
}

// uploadTestReadCloser proves the use case leaves body ownership to the HTTP
// handler.
type uploadTestReadCloser struct {
	io.Reader
	closed bool
}

func (r *uploadTestReadCloser) Close() error {
	r.closed = true
	return nil
}

func TestUploadAttachmentUsecaseSuccess(t *testing.T) {
	h := newUploadHarness()

	var (
		putCalled  bool
		gotKey     string
		gotOpts    attachmentout.PutOptions
		stored     []byte
		gotHead    []byte
		gotDecl    string
		saved      *domain.Attachment
		accessArgs [5]string
		action     attachmentout.Action
	)
	h.detector.DetectFunc = func(head []byte, declared string) (domain.ContentType, error) {
		gotHead = append([]byte(nil), head...)
		gotDecl = declared
		return domain.ContentTypePDF, nil
	}
	h.access.CanAccessFunc = func(_ context.Context, refName, refID, actorID, actorRole string, a attachmentout.Action) error {
		accessArgs = [5]string{refName, refID, actorID, actorRole, string(a)}
		action = a
		return nil
	}
	h.storage.PutFunc = func(_ context.Context, key string, r io.Reader, opts attachmentout.PutOptions) (attachmentout.ObjectInfo, error) {
		putCalled = true
		gotKey, gotOpts = key, opts
		var err error
		stored, err = io.ReadAll(r)
		if err != nil {
			return attachmentout.ObjectInfo{}, err
		}
		sum := sha256.Sum256(stored)
		return attachmentout.ObjectInfo{
			Key:      key,
			Size:     int64(len(stored)),
			Checksum: hex.EncodeToString(sum[:]),
		}, nil
	}
	h.repo.SaveFunc = func(_ context.Context, a *domain.Attachment) error {
		if !putCalled {
			t.Error("metadata was saved before bytes were stored")
		}
		saved = a
		return nil
	}

	bodyReader := &uploadTestReadCloser{Reader: bytes.NewReader(uploadTestBody)}
	cmd := validUploadCommand(uploadTestBody)
	cmd.Body = bodyReader

	attachment, err := h.uc.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if bodyReader.closed {
		t.Error("use case closed the caller's body")
	}
	if !bytes.Equal(gotHead, uploadTestBody) || gotDecl != "" {
		t.Errorf("Detect(head=%q, declared=%q), want body and empty declaration", gotHead, gotDecl)
	}
	if action != attachmentout.ActionUpload || accessArgs != [5]string{"requisition", attachmentmocks.DefaultRefID, attachmentmocks.DefaultActorID, attachmentmocks.DefaultActorRole, "upload"} {
		t.Errorf("access check = %v/%q", accessArgs, action)
	}
	if !bytes.Equal(stored, uploadTestBody) {
		t.Errorf("stored bytes = %q, want %q", stored, uploadTestBody)
	}
	if gotKey != attachmentmocks.DefaultStorageKey {
		t.Errorf("storage key = %q, want %q", gotKey, attachmentmocks.DefaultStorageKey)
	}

	wantHash := sha256.Sum256(uploadTestBody)
	wantChecksum := hex.EncodeToString(wantHash[:])
	if gotOpts.ContentType != domain.ContentTypePDF.String() {
		t.Errorf("Put content type = %q, want %q", gotOpts.ContentType, domain.ContentTypePDF)
	}
	if gotOpts.Size != int64(len(uploadTestBody)) {
		t.Errorf("Put declared size = %d, want %d", gotOpts.Size, len(uploadTestBody))
	}

	if saved == nil || attachment == nil || saved.ID() != attachment.ID() {
		t.Fatal("expected the saved attachment to be returned")
	}
	if attachment.ID() != attachmentmocks.DefaultAttachmentID {
		t.Errorf("attachment ID = %q", attachment.ID())
	}
	if attachment.StorageKey() != gotKey {
		t.Errorf("attachment key = %q, want %q", attachment.StorageKey(), gotKey)
	}
	if attachment.Checksum() != wantChecksum {
		t.Errorf("attachment checksum = %q, want %q", attachment.Checksum(), wantChecksum)
	}
	if attachment.FileSize() != int64(len(uploadTestBody)) {
		t.Errorf("attachment size = %d, want %d", attachment.FileSize(), len(uploadTestBody))
	}
	if attachment.ContentType() != domain.ContentTypePDF {
		t.Errorf("attachment content type = %q", attachment.ContentType())
	}
}

func TestUploadAttachmentUsecaseUnsupportedContentType(t *testing.T) {
	tests := []struct {
		name     string
		detect   func([]byte, string) (domain.ContentType, error)
		wantErr  error
		wantSame bool
	}{
		{
			name: "plain detector error is normalised",
			detect: func([]byte, string) (domain.ContentType, error) {
				return "", errors.New("magic bytes rejected")
			},
			wantErr: ErrUnsupportedContentType,
		},
		{
			name: "custom error keeps its status",
			detect: func([]byte, string) (domain.ContentType, error) {
				return "", customerrors.NewCustomError("only PDFs", http.StatusUnsupportedMediaType)
			},
			wantSame: true,
		},
		{
			name: "invalid type without error is rejected",
			detect: func([]byte, string) (domain.ContentType, error) {
				return domain.ContentType("text/html"), nil
			},
			wantErr: ErrUnsupportedContentType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newUploadHarness()
			h.detector.DetectFunc = tt.detect
			putCalled := false
			h.storage.PutFunc = func(context.Context, string, io.Reader, attachmentout.PutOptions) (attachmentout.ObjectInfo, error) {
				putCalled = true
				return attachmentout.ObjectInfo{}, nil
			}

			_, err := h.uc.Execute(context.Background(), validUploadCommand(uploadTestBody))
			if err == nil {
				t.Fatal("expected an error")
			}
			if tt.wantSame {
				if ce, ok := err.(*customerrors.CustomError); !ok || ce.Code != http.StatusUnsupportedMediaType {
					t.Fatalf("error = %v (%T), want the original 415", err, err)
				}
			} else if err != tt.wantErr {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if putCalled {
				t.Error("bytes were stored for an unsupported type")
			}
		})
	}
}

func TestUploadAttachmentUsecaseOversize(t *testing.T) {
	t.Run("declared size", func(t *testing.T) {
		h := newUploadHarness()
		cmd := validUploadCommand(uploadTestBody)
		cmd.Size = uploadTestMaxSize + 1
		detectCalled, putCalled := false, false
		h.detector.DetectFunc = func([]byte, string) (domain.ContentType, error) {
			detectCalled = true
			return domain.ContentTypePDF, nil
		}
		h.storage.PutFunc = func(context.Context, string, io.Reader, attachmentout.PutOptions) (attachmentout.ObjectInfo, error) {
			putCalled = true
			return attachmentout.ObjectInfo{}, nil
		}

		_, err := h.uc.Execute(context.Background(), cmd)
		if err != ErrAttachmentTooLarge {
			t.Fatalf("error = %v, want ErrAttachmentTooLarge", err)
		}
		if detectCalled || putCalled {
			t.Error("oversize declared body was detected or stored")
		}
	})

	t.Run("body lies about its size", func(t *testing.T) {
		h := newUploadHarness()
		body := bytes.Repeat([]byte("a"), int(uploadTestMaxSize)+1)
		putCalled, saveCalled := false, false
		h.storage.PutFunc = func(_ context.Context, key string, r io.Reader, _ attachmentout.PutOptions) (attachmentout.ObjectInfo, error) {
			putCalled = true
			n, err := io.Copy(io.Discard, r)
			return attachmentout.ObjectInfo{Key: key, Size: n}, err
		}
		h.repo.SaveFunc = func(context.Context, *domain.Attachment) error {
			saveCalled = true
			return nil
		}

		cmd := validUploadCommand(body)
		cmd.Size = 1 // the client claims a tiny file but streams an oversize body
		_, err := h.uc.Execute(context.Background(), cmd)
		if err != ErrAttachmentTooLarge {
			t.Fatalf("error = %v, want ErrAttachmentTooLarge", err)
		}
		if !putCalled {
			t.Error("storage was not asked to stream the body")
		}
		if saveCalled {
			t.Error("oversize bytes were saved as metadata")
		}
	})

	t.Run("storage reports an oversize object", func(t *testing.T) {
		h := newUploadHarness()
		deletedKey := ""
		h.storage.PutFunc = func(_ context.Context, key string, _ io.Reader, _ attachmentout.PutOptions) (attachmentout.ObjectInfo, error) {
			return attachmentout.ObjectInfo{Key: key, Size: uploadTestMaxSize + 1}, nil
		}
		h.storage.DeleteFunc = func(_ context.Context, key string) error {
			deletedKey = key
			return nil
		}

		_, err := h.uc.Execute(context.Background(), validUploadCommand(uploadTestBody))
		if err != ErrAttachmentTooLarge {
			t.Fatalf("error = %v, want ErrAttachmentTooLarge", err)
		}
		if deletedKey != attachmentmocks.DefaultStorageKey {
			t.Errorf("cleanup key = %q, want %q", deletedKey, attachmentmocks.DefaultStorageKey)
		}
	})
}

func TestUploadAttachmentUsecaseOverCountAndAccess(t *testing.T) {
	t.Run("parent is full", func(t *testing.T) {
		h := newUploadHarness()
		h.repo.CountByRefFunc = func(context.Context, string, string) (int, error) {
			return uploadTestMaxPerRef, nil
		}
		putCalled := false
		h.storage.PutFunc = func(context.Context, string, io.Reader, attachmentout.PutOptions) (attachmentout.ObjectInfo, error) {
			putCalled = true
			return attachmentout.ObjectInfo{}, nil
		}

		_, err := h.uc.Execute(context.Background(), validUploadCommand(uploadTestBody))
		if err != ErrAttachmentLimitReached {
			t.Fatalf("error = %v, want ErrAttachmentLimitReached", err)
		}
		if putCalled {
			t.Error("bytes were stored for a full parent")
		}
	})

	t.Run("access denied", func(t *testing.T) {
		h := newUploadHarness()
		denied := customerrors.NewForbiddenError("no access to parent")
		h.access.CanAccessFunc = func(context.Context, string, string, string, string, attachmentout.Action) error {
			return denied
		}
		countCalled, putCalled := false, false
		h.repo.CountByRefFunc = func(context.Context, string, string) (int, error) {
			countCalled = true
			return 0, nil
		}
		h.storage.PutFunc = func(context.Context, string, io.Reader, attachmentout.PutOptions) (attachmentout.ObjectInfo, error) {
			putCalled = true
			return attachmentout.ObjectInfo{}, nil
		}

		_, err := h.uc.Execute(context.Background(), validUploadCommand(uploadTestBody))
		if !errors.Is(err, denied) {
			t.Fatalf("error = %v, want access error", err)
		}
		if countCalled || putCalled {
			t.Error("work was done after access was denied")
		}
	})
}

func TestUploadAttachmentUsecaseMetadataFailureDeletesObject(t *testing.T) {
	h := newUploadHarness()
	saveErr := errors.New("database unavailable")
	deletedKey := ""
	h.repo.SaveFunc = func(context.Context, *domain.Attachment) error { return saveErr }
	h.storage.DeleteFunc = func(_ context.Context, key string) error {
		deletedKey = key
		return nil
	}

	_, err := h.uc.Execute(context.Background(), validUploadCommand(uploadTestBody))
	if !errors.Is(err, saveErr) {
		t.Fatalf("error = %v, want the original save error", err)
	}
	if deletedKey != attachmentmocks.DefaultStorageKey {
		t.Errorf("cleanup key = %q, want %q", deletedKey, attachmentmocks.DefaultStorageKey)
	}

	t.Run("transaction failure also cleans up", func(t *testing.T) {
		h := newUploadHarness()
		txErr := errors.New("transaction rolled back")
		h.tx.WithinTransactionFunc = func(context.Context, func(context.Context) error) error { return txErr }
		deletedKey := ""
		h.storage.DeleteFunc = func(_ context.Context, key string) error {
			deletedKey = key
			return nil
		}

		_, err := h.uc.Execute(context.Background(), validUploadCommand(uploadTestBody))
		if !errors.Is(err, txErr) {
			t.Fatalf("error = %v, want txErr", err)
		}
		if deletedKey != attachmentmocks.DefaultStorageKey {
			t.Errorf("cleanup key = %q, want %q", deletedKey, attachmentmocks.DefaultStorageKey)
		}
	})
}

func TestUploadAttachmentUsecaseStorageFailureLeavesNoRow(t *testing.T) {
	h := newUploadHarness()
	storageErr := errors.New("write failed")
	h.storage.PutFunc = func(context.Context, string, io.Reader, attachmentout.PutOptions) (attachmentout.ObjectInfo, error) {
		return attachmentout.ObjectInfo{}, storageErr
	}
	saveCalled := false
	h.repo.SaveFunc = func(context.Context, *domain.Attachment) error {
		saveCalled = true
		return nil
	}

	_, err := h.uc.Execute(context.Background(), validUploadCommand(uploadTestBody))
	if !errors.Is(err, storageErr) {
		t.Fatalf("error = %v, want storageErr", err)
	}
	if saveCalled {
		t.Error("metadata was saved after a failed object write")
	}
}

func TestUploadAttachmentUsecaseValidation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*attachmentin.UploadAttachmentCommand)
		wantErr error
	}{
		{"missing actor", func(c *attachmentin.UploadAttachmentCommand) { c.ActorID = "" }, ErrAttachmentActorRequired},
		{"missing role", func(c *attachmentin.UploadAttachmentCommand) { c.ActorRole = "" }, ErrAttachmentActorRequired},
		{"bad ref name", func(c *attachmentin.UploadAttachmentCommand) { c.RefName = "user" }, domain.ErrInvalidAttachmentRefName},
		{"bad ref ID", func(c *attachmentin.UploadAttachmentCommand) { c.RefID = "att-1234" }, domain.ErrInvalidAttachmentRefID},
		{"missing file name", func(c *attachmentin.UploadAttachmentCommand) { c.FileName = "" }, domain.ErrInvalidAttachmentFileName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newUploadHarness()
			cmd := validUploadCommand(uploadTestBody)
			tt.mutate(&cmd)
			if _, err := h.uc.Execute(context.Background(), cmd); err != tt.wantErr {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}

	t.Run("cancelled context", func(t *testing.T) {
		h := newUploadHarness()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := h.uc.Execute(ctx, validUploadCommand(uploadTestBody)); !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want context.Canceled", err)
		}
	})
}

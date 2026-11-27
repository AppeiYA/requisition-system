package usecase

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentmocks "github.com/AppeiYA/requisition-system/internal/attachment/mocks"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
)

const (
	stageTestMaxSize = int64(64)
	stageTestTTL     = time.Hour
)

var stageTestBody = []byte("%PDF-1.4\nstaged body\n")

type stageHarness struct {
	repo     *attachmentmocks.MockStagedUploadRepository
	storage  *attachmentmocks.MockFileStorage
	detector *attachmentmocks.MockContentTypeDetector
	idGen    *attachmentmocks.MockIDGenerator
	tx       *attachmentmocks.MockTransactionManager
	uc       *StageUploadUsecase
}

func newStageHarness() *stageHarness {
	h := &stageHarness{
		repo:     attachmentmocks.NewMockStagedUploadRepository(),
		storage:  attachmentmocks.NewMockFileStorage(),
		detector: attachmentmocks.NewMockContentTypeDetector(),
		idGen:    attachmentmocks.NewMockIDGenerator(),
		tx:       attachmentmocks.NewMockTransactionManager(),
	}
	h.idGen.GenerateIDFunc = func(prefix string) (string, error) {
		return "upl-1234", nil
	}
	h.uc = NewStageUploadUsecase(
		h.repo, h.storage, h.detector, h.idGen, h.tx, stageTestMaxSize, stageTestTTL,
	)
	h.uc.now = func() time.Time { return time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC) }
	return h
}

func validStageCommand(body []byte) attachmentin.StageUploadCommand {
	return attachmentin.StageUploadCommand{
		ActorID:  attachmentmocks.DefaultActorID,
		FileName: "quote.pdf",
		Body:     bytes.NewReader(body),
		Size:     int64(len(body)),
	}
}

func TestStageUploadUsecaseSuccess(t *testing.T) {
	h := newStageHarness()
	var (
		gotKey  string
		stored  []byte
		saved   *domain.StagedUpload
		gotOpts attachmentout.PutOptions
	)
	h.storage.PutFunc = func(_ context.Context, key string, r io.Reader, opts attachmentout.PutOptions) (attachmentout.ObjectInfo, error) {
		gotKey = key
		gotOpts = opts
		var err error
		stored, err = io.ReadAll(r)
		sum := sha256.Sum256(stored)
		return attachmentout.ObjectInfo{
			Key:      key,
			Size:     int64(len(stored)),
			Checksum: hex.EncodeToString(sum[:]),
		}, err
	}
	h.repo.SaveFunc = func(_ context.Context, upload *domain.StagedUpload) error {
		saved = upload
		return nil
	}

	upload, err := h.uc.Execute(context.Background(), validStageCommand(stageTestBody))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	wantHash := sha256.Sum256(stageTestBody)
	wantChecksum := hex.EncodeToString(wantHash[:])

	if gotKey != attachmentmocks.DefaultStagingPath {
		t.Errorf("storage key = %q, want %q", gotKey, attachmentmocks.DefaultStagingPath)
	}
	if !bytes.Equal(stored, stageTestBody) {
		t.Errorf("stored bytes = %q, want %q", stored, stageTestBody)
	}
	if gotOpts.ContentType != domain.ContentTypePDF.String() || gotOpts.Size != int64(len(stageTestBody)) {
		t.Errorf("Put options = %+v", gotOpts)
	}
	if saved == nil || saved != upload {
		t.Fatal("expected the saved staged upload to be returned")
	}
	if upload.ID() != attachmentmocks.DefaultStagedID {
		t.Errorf("ID = %q, want %q", upload.ID(), attachmentmocks.DefaultStagedID)
	}
	if upload.StoragePath() != gotKey {
		t.Errorf("StoragePath() = %q, want %q", upload.StoragePath(), gotKey)
	}
	if upload.Checksum() != wantChecksum {
		t.Errorf("Checksum() = %q, want %q", upload.Checksum(), wantChecksum)
	}
	if upload.Status() != domain.StagedUploadStatusPending {
		t.Errorf("Status() = %q, want pending", upload.Status())
	}
	if !upload.ExpiresAt().Equal(upload.CreatedAt().Add(stageTestTTL)) {
		t.Errorf("ExpiresAt() = %v, want %v", upload.ExpiresAt(), upload.CreatedAt().Add(stageTestTTL))
	}
}

func TestStageUploadUsecaseEnforcesQuota(t *testing.T) {
	h := newStageHarness()
	h.repo.CountPendingByUserFunc = func(context.Context, string, time.Time) (int, int64, error) {
		return 5, 1024, nil
	}
	uc := NewStageUploadUsecase(
		h.repo, h.storage, h.detector, h.idGen, h.tx,
		10<<20, 24*time.Hour, 5, 0,
	)

	_, err := uc.Execute(context.Background(), validStageCommand(stageTestBody))
	if err != ErrStagedUploadQuotaExceeded {
		t.Fatalf("Execute() error = %v, want ErrStagedUploadQuotaExceeded", err)
	}
}

func TestStageUploadUsecaseFailures(t *testing.T) {
	t.Run("missing actor", func(t *testing.T) {
		h := newStageHarness()
		cmd := validStageCommand(stageTestBody)
		cmd.ActorID = ""
		if _, err := h.uc.Execute(context.Background(), cmd); !errors.Is(err, ErrAttachmentActorRequired) {
			t.Fatalf("error = %v, want ErrAttachmentActorRequired", err)
		}
	})

	t.Run("oversize declared body", func(t *testing.T) {
		h := newStageHarness()
		cmd := validStageCommand(stageTestBody)
		cmd.Size = stageTestMaxSize + 1
		if _, err := h.uc.Execute(context.Background(), cmd); !errors.Is(err, ErrAttachmentTooLarge) {
			t.Fatalf("error = %v, want ErrAttachmentTooLarge", err)
		}
	})

	t.Run("metadata failure deletes staged object", func(t *testing.T) {
		h := newStageHarness()
		saveErr := errors.New("database unavailable")
		deletedKey := ""
		h.repo.SaveFunc = func(context.Context, *domain.StagedUpload) error { return saveErr }
		h.storage.DeleteFunc = func(_ context.Context, key string) error {
			deletedKey = key
			return nil
		}

		if _, err := h.uc.Execute(context.Background(), validStageCommand(stageTestBody)); !errors.Is(err, saveErr) {
			t.Fatalf("error = %v, want save error", err)
		}
		if deletedKey != attachmentmocks.DefaultStagingPath {
			t.Errorf("cleanup key = %q, want %q", deletedKey, attachmentmocks.DefaultStagingPath)
		}
	})

	t.Run("cancelled context", func(t *testing.T) {
		h := newStageHarness()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := h.uc.Execute(ctx, validStageCommand(stageTestBody)); !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want context.Canceled", err)
		}
	})
}

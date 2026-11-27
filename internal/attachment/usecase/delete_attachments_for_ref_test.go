package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentmocks "github.com/AppeiYA/requisition-system/internal/attachment/mocks"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
)

func forRefTestCommand() attachmentin.DeleteForRefCommand {
	return attachmentin.DeleteForRefCommand{
		RefName: attachmentmocks.DefaultRefName,
		RefID:   attachmentmocks.DefaultRefID,
	}
}

func TestDeleteAttachmentsForRefUsecaseDeletesRowsThenObjects(t *testing.T) {
	repo := attachmentmocks.NewMockAttachmentRepository()
	storage := attachmentmocks.NewMockFileStorage()
	tx := attachmentmocks.NewMockTransactionManager()
	events := []string{}
	keys := []string{"requisition/req-1234/att-1", "requisition/req-1234/att-2"}

	repo.DeleteByRefFunc = func(_ context.Context, refName, refID string) ([]string, error) {
		events = append(events, "rows")
		if refName != attachmentmocks.DefaultRefName || refID != attachmentmocks.DefaultRefID {
			t.Errorf("DeleteByRef(%q, %q)", refName, refID)
		}
		return keys, nil
	}
	tx.WithinTransactionFunc = func(ctx context.Context, fn func(context.Context) error) error {
		events = append(events, "tx")
		return fn(ctx)
	}
	deleted := []string{}
	storage.DeleteFunc = func(_ context.Context, key string) error {
		events = append(events, "object")
		deleted = append(deleted, key)
		return nil
	}

	count, err := NewDeleteAttachmentsForRefUsecase(repo, storage, tx).Execute(context.Background(), forRefTestCommand())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if count != len(keys) {
		t.Errorf("count = %d, want %d", count, len(keys))
	}
	if len(deleted) != len(keys) || deleted[0] != keys[0] || deleted[1] != keys[1] {
		t.Errorf("deleted keys = %v, want %v", deleted, keys)
	}
	if len(events) != 4 || events[0] != "tx" || events[1] != "rows" || events[2] != "object" || events[3] != "object" {
		t.Errorf("events = %v, want [tx rows object object]", events)
	}
}

func TestDeleteAttachmentsForRefUsecaseFailureAndIdempotence(t *testing.T) {
	t.Run("invalid ref", func(t *testing.T) {
		cmd := forRefTestCommand()
		cmd.RefID = "att-1234"
		count, err := NewDeleteAttachmentsForRefUsecase(
			attachmentmocks.NewMockAttachmentRepository(),
			attachmentmocks.NewMockFileStorage(),
			attachmentmocks.NewMockTransactionManager(),
		).Execute(context.Background(), cmd)
		if err != domain.ErrInvalidAttachmentRefID {
			t.Fatalf("error = %v, want ErrInvalidAttachmentRefID", err)
		}
		if count != 0 {
			t.Errorf("count = %d, want 0", count)
		}
	})

	t.Run("repository failure leaves objects alone", func(t *testing.T) {
		repo := attachmentmocks.NewMockAttachmentRepository()
		repoErr := errors.New("query failed")
		repo.DeleteByRefFunc = func(context.Context, string, string) ([]string, error) {
			return nil, repoErr
		}
		storageCalled := false
		storage := attachmentmocks.NewMockFileStorage()
		storage.DeleteFunc = func(context.Context, string) error {
			storageCalled = true
			return nil
		}
		_, err := NewDeleteAttachmentsForRefUsecase(repo, storage, attachmentmocks.NewMockTransactionManager()).
			Execute(context.Background(), forRefTestCommand())
		if !errors.Is(err, repoErr) {
			t.Fatalf("error = %v, want repoErr", err)
		}
		if storageCalled {
			t.Error("objects were deleted after a repository failure")
		}
	})

	t.Run("storage failure is tolerated and empty refs are idempotent", func(t *testing.T) {
		repo := attachmentmocks.NewMockAttachmentRepository()
		repo.DeleteByRefFunc = func(context.Context, string, string) ([]string, error) {
			return []string{attachmentmocks.DefaultStorageKey}, nil
		}
		storage := attachmentmocks.NewMockFileStorage()
		storage.DeleteFunc = func(context.Context, string) error {
			return errors.New("object store unavailable")
		}
		uc := NewDeleteAttachmentsForRefUsecase(repo, storage, attachmentmocks.NewMockTransactionManager())
		if count, err := uc.Execute(context.Background(), forRefTestCommand()); err != nil || count != 1 {
			t.Fatalf("Execute() = (%d, %v), want (1, nil)", count, err)
		}
		repo.DeleteByRefFunc = func(context.Context, string, string) ([]string, error) {
			return nil, nil
		}
		if count, err := uc.Execute(context.Background(), forRefTestCommand()); err != nil || count != 0 {
			t.Fatalf("second Execute() = (%d, %v), want (0, nil)", count, err)
		}
	})
}

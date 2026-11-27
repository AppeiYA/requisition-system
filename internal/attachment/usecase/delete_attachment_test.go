package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentmocks "github.com/AppeiYA/requisition-system/internal/attachment/mocks"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

func deleteTestCommand() attachmentin.DeleteAttachmentCommand {
	return attachmentin.DeleteAttachmentCommand{
		AttachmentID: attachmentmocks.DefaultAttachmentID,
		ActorID:      attachmentmocks.DefaultActorID,
		ActorRole:    attachmentmocks.DefaultActorRole,
	}
}

func TestDeleteAttachmentUsecaseDeletesRowBeforeObject(t *testing.T) {
	repo := attachmentmocks.NewMockAttachmentRepository()
	storage := attachmentmocks.NewMockFileStorage()
	access := attachmentmocks.NewMockRefAccessChecker()
	events := []string{}
	var action attachmentout.Action

	repo.DeleteFunc = func(_ context.Context, id string) error {
		if id != attachmentmocks.DefaultAttachmentID {
			t.Errorf("Delete(%q)", id)
		}
		events = append(events, "row")
		return nil
	}
	storage.DeleteFunc = func(_ context.Context, key string) error {
		if key != attachmentmocks.DefaultStorageKey {
			t.Errorf("storage.Delete(%q)", key)
		}
		events = append(events, "object")
		return nil
	}
	access.CanAccessFunc = func(_ context.Context, refName, refID, actorID, actorRole string, a attachmentout.Action) error {
		action = a
		if refName != attachmentmocks.DefaultRefName || refID != attachmentmocks.DefaultRefID ||
			actorID != attachmentmocks.DefaultActorID || actorRole != attachmentmocks.DefaultActorRole {
			t.Errorf("unexpected access arguments: %s/%s/%s/%s", refName, refID, actorID, actorRole)
		}
		return nil
	}

	if err := NewDeleteAttachmentUsecase(repo, storage, access, attachmentmocks.NewMockTransactionManager()).
		Execute(context.Background(), deleteTestCommand()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(events) != 2 || events[0] != "row" || events[1] != "object" {
		t.Errorf("events = %v, want [row object]", events)
	}
	if action != attachmentout.ActionDelete {
		t.Errorf("access action = %q, want delete", action)
	}
}

func TestDeleteAttachmentUsecaseToleratesStorageFailure(t *testing.T) {
	storage := attachmentmocks.NewMockFileStorage()
	repo := attachmentmocks.NewMockAttachmentRepository()
	storage.DeleteFunc = func(context.Context, string) error {
		return errors.New("bucket unavailable")
	}

	if err := NewDeleteAttachmentUsecase(repo, storage, attachmentmocks.NewMockRefAccessChecker(), attachmentmocks.NewMockTransactionManager()).
		Execute(context.Background(), deleteTestCommand()); err != nil {
		t.Fatalf("storage failure must not fail the request, got %v", err)
	}
}

func TestDeleteAttachmentUsecaseRefusesDeniedMissingAndFailedTransaction(t *testing.T) {
	t.Run("access denied", func(t *testing.T) {
		access := attachmentmocks.NewMockRefAccessChecker()
		denied := customerrors.NewForbiddenError("denied")
		access.CanAccessFunc = func(context.Context, string, string, string, string, attachmentout.Action) error {
			return denied
		}
		storageCalled := false
		storage := attachmentmocks.NewMockFileStorage()
		storage.DeleteFunc = func(context.Context, string) error {
			storageCalled = true
			return nil
		}
		err := NewDeleteAttachmentUsecase(attachmentmocks.NewMockAttachmentRepository(), storage, access, attachmentmocks.NewMockTransactionManager()).
			Execute(context.Background(), deleteTestCommand())
		if !errors.Is(err, denied) {
			t.Fatalf("error = %v, want access error", err)
		}
		if storageCalled {
			t.Error("object was deleted after access was denied")
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := attachmentmocks.NewMockAttachmentRepository()
		repo.GetByIDFunc = func(context.Context, string) (*domain.Attachment, error) {
			return nil, customerrors.NewNotFoundError("row missing")
		}
		err := NewDeleteAttachmentUsecase(repo, attachmentmocks.NewMockFileStorage(), attachmentmocks.NewMockRefAccessChecker(), attachmentmocks.NewMockTransactionManager()).
			Execute(context.Background(), deleteTestCommand())
		if err != ErrAttachmentNotFound {
			t.Fatalf("error = %v, want ErrAttachmentNotFound", err)
		}
	})

	t.Run("transaction failure", func(t *testing.T) {
		txErr := errors.New("transaction failed")
		tx := attachmentmocks.NewMockTransactionManager()
		tx.WithinTransactionFunc = func(context.Context, func(context.Context) error) error { return txErr }
		storageCalled := false
		storage := attachmentmocks.NewMockFileStorage()
		storage.DeleteFunc = func(context.Context, string) error {
			storageCalled = true
			return nil
		}
		err := NewDeleteAttachmentUsecase(attachmentmocks.NewMockAttachmentRepository(), storage, attachmentmocks.NewMockRefAccessChecker(), tx).
			Execute(context.Background(), deleteTestCommand())
		if !errors.Is(err, txErr) {
			t.Fatalf("error = %v, want txErr", err)
		}
		if storageCalled {
			t.Error("object was deleted before the row transaction committed")
		}
	})
}

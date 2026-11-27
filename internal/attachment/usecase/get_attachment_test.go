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

func getTestQuery() attachmentin.GetAttachmentQuery {
	return attachmentin.GetAttachmentQuery{
		AttachmentID: attachmentmocks.DefaultAttachmentID,
		ActorID:      attachmentmocks.DefaultActorID,
		ActorRole:    attachmentmocks.DefaultActorRole,
	}
}

func TestGetAttachmentUsecaseReturnsCheckedMetadata(t *testing.T) {
	repo := attachmentmocks.NewMockAttachmentRepository()
	access := attachmentmocks.NewMockRefAccessChecker()
	var action attachmentout.Action
	access.CanAccessFunc = func(_ context.Context, _, _, _, _ string, a attachmentout.Action) error {
		action = a
		return nil
	}

	attachment, err := NewGetAttachmentUsecase(repo, access).Execute(context.Background(), getTestQuery())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if attachment == nil || attachment.ID() != attachmentmocks.DefaultAttachmentID {
		t.Fatalf("unexpected attachment: %v", attachment)
	}
	if action != attachmentout.ActionRead {
		t.Errorf("access action = %q, want read", action)
	}
}

func TestGetAttachmentUsecaseNotFoundAndDenied(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		repo := attachmentmocks.NewMockAttachmentRepository()
		repo.GetByIDFunc = func(context.Context, string) (*domain.Attachment, error) { return nil, nil }
		_, err := NewGetAttachmentUsecase(repo, attachmentmocks.NewMockRefAccessChecker()).Execute(context.Background(), getTestQuery())
		if err != ErrAttachmentNotFound {
			t.Fatalf("error = %v, want ErrAttachmentNotFound", err)
		}
	})

	t.Run("denied", func(t *testing.T) {
		access := attachmentmocks.NewMockRefAccessChecker()
		denied := customerrors.NewForbiddenError("denied")
		access.CanAccessFunc = func(context.Context, string, string, string, string, attachmentout.Action) error {
			return denied
		}
		_, err := NewGetAttachmentUsecase(attachmentmocks.NewMockAttachmentRepository(), access).
			Execute(context.Background(), getTestQuery())
		if !errors.Is(err, denied) {
			t.Fatalf("error = %v, want access error", err)
		}
	})
}

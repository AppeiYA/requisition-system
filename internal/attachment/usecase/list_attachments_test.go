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
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/pagination"
)

func listTestQuery() attachmentin.ListAttachmentsQuery {
	return attachmentin.ListAttachmentsQuery{
		ActorID:   attachmentmocks.DefaultActorID,
		ActorRole: attachmentmocks.DefaultActorRole,
		RefName:   attachmentmocks.DefaultRefName,
		RefID:     attachmentmocks.DefaultRefID,
		Limit:     0,
		Offset:    -1,
	}
}

func TestListAttachmentsUsecaseClampsAndChecksAccess(t *testing.T) {
	repo := attachmentmocks.NewMockAttachmentRepository()
	access := attachmentmocks.NewMockRefAccessChecker()

	var (
		accessRefName, accessRefID string
		listRefName, listRefID     string
		gotLimit, gotOffset        int
		action                     attachmentout.Action
	)
	access.CanAccessFunc = func(_ context.Context, refName, refID, _, _ string, a attachmentout.Action) error {
		accessRefName, accessRefID, action = refName, refID, a
		return nil
	}
	repo.ListByRefFunc = func(_ context.Context, refName, refID string, limit, offset int) ([]*domain.Attachment, int, error) {
		listRefName, listRefID = refName, refID
		gotLimit, gotOffset = limit, offset
		return []*domain.Attachment{attachmentRecordWithSize(16)}, 999, nil
	}

	attachments, err := NewListAttachmentsUsecase(repo, access).Execute(context.Background(), listTestQuery())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(attachments) != 1 || attachments[0].ID() != attachmentmocks.DefaultAttachmentID {
		t.Fatalf("unexpected attachments: %v", attachments)
	}
	wantRef := attachmentmocks.DefaultRefName + "/" + attachmentmocks.DefaultRefID
	if accessRefName+"/"+accessRefID != wantRef {
		t.Errorf("access check for %q/%q, want %s", accessRefName, accessRefID, wantRef)
	}
	if listRefName+"/"+listRefID != wantRef {
		t.Errorf("ListByRef(%q, %q), want %s", listRefName, listRefID, wantRef)
	}
	if gotLimit != pagination.MaxLimit || gotOffset != 0 {
		t.Errorf("ListByRef(limit=%d, offset=%d), want %d/0", gotLimit, gotOffset, pagination.MaxLimit)
	}
	if action != attachmentout.ActionRead {
		t.Errorf("access action = %q, want read", action)
	}
}

func TestListAttachmentsUsecaseValidationAndDenial(t *testing.T) {
	t.Run("invalid ref", func(t *testing.T) {
		query := listTestQuery()
		query.RefID = "att-1234"
		_, err := NewListAttachmentsUsecase(attachmentmocks.NewMockAttachmentRepository(), attachmentmocks.NewMockRefAccessChecker()).
			Execute(context.Background(), query)
		if err != domain.ErrInvalidAttachmentRefID {
			t.Fatalf("error = %v, want ErrInvalidAttachmentRefID", err)
		}
	})

	t.Run("denied", func(t *testing.T) {
		access := attachmentmocks.NewMockRefAccessChecker()
		denied := customerrors.NewForbiddenError("denied")
		access.CanAccessFunc = func(context.Context, string, string, string, string, attachmentout.Action) error {
			return denied
		}
		called := false
		repo := attachmentmocks.NewMockAttachmentRepository()
		repo.ListByRefFunc = func(context.Context, string, string, int, int) ([]*domain.Attachment, int, error) {
			called = true
			return nil, 0, nil
		}
		_, err := NewListAttachmentsUsecase(repo, access).Execute(context.Background(), listTestQuery())
		if !errors.Is(err, denied) {
			t.Fatalf("error = %v, want access error", err)
		}
		if called {
			t.Error("repository was queried after access was denied")
		}
	})
}

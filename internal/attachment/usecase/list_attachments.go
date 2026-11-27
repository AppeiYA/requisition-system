package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/pagination"
)

var _ attachmentin.ListAttachmentsUsecaseInt = (*ListAttachmentsUsecase)(nil)

// ListAttachmentsUsecase lists metadata for one parent resource.
type ListAttachmentsUsecase struct {
	repo   attachmentout.AttachmentRepository
	access attachmentout.RefAccessChecker
}

func NewListAttachmentsUsecase(
	repo attachmentout.AttachmentRepository,
	access attachmentout.RefAccessChecker,
) *ListAttachmentsUsecase {
	return &ListAttachmentsUsecase{repo: repo, access: access}
}

func (uc *ListAttachmentsUsecase) Execute(ctx context.Context, query attachmentin.ListAttachmentsQuery) ([]*domain.Attachment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ref, err := domain.NewAttachmentRef(domain.AttachmentRefName(query.RefName), query.RefID)
	if err != nil {
		return nil, err
	}
	refName, refID := ref.RefName().String(), ref.ID()

	if err := uc.access.CanAccess(ctx, refName, refID, query.ActorID, query.ActorRole, attachmentout.ActionRead); err != nil {
		return nil, err
	}

	limit, offset := pagination.Clamp(query.Limit, query.Offset)
	attachments, _, err := uc.repo.ListByRef(ctx, refName, refID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Pending rows are visible to the owner through the requisition itself,
	// but metadata listing only shows files whose bytes are downloadable.
	ready := make([]*domain.Attachment, 0, len(attachments))
	for _, attachment := range attachments {
		if attachment.Status() == domain.AttachmentStatusReady {
			ready = append(ready, attachment)
		}
	}
	return ready, nil
}

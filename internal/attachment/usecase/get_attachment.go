package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
)

var _ attachmentin.GetAttachmentUsecaseInt = (*GetAttachmentUsecase)(nil)

// GetAttachmentUsecase returns attachment metadata only; bytes are served by
// DownloadAttachmentUsecase.
type GetAttachmentUsecase struct {
	repo   attachmentout.AttachmentRepository
	access attachmentout.RefAccessChecker
}

func NewGetAttachmentUsecase(
	repo attachmentout.AttachmentRepository,
	access attachmentout.RefAccessChecker,
) *GetAttachmentUsecase {
	return &GetAttachmentUsecase{repo: repo, access: access}
}

func (uc *GetAttachmentUsecase) Execute(ctx context.Context, query attachmentin.GetAttachmentQuery) (*domain.Attachment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	attachment, err := uc.repo.GetByID(ctx, query.AttachmentID)
	if err != nil {
		return nil, attachmentRecordError(err, nil)
	}
	if err := attachmentRecordError(nil, attachment); err != nil {
		return nil, err
	}

	ref := attachment.Ref()
	if err := uc.access.CanAccess(ctx, ref.RefName().String(), ref.ID(), query.ActorID, query.ActorRole, attachmentout.ActionRead); err != nil {
		return nil, err
	}
	if attachment.Status() != domain.AttachmentStatusReady {
		return nil, domain.ErrAttachmentNotReady
	}

	return attachment, nil
}

package usecase

import (
	"context"
	"log/slog"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
)

var _ attachmentin.DeleteAttachmentUsecaseInt = (*DeleteAttachmentUsecase)(nil)

// DeleteAttachmentUsecase removes metadata inside a transaction and then the
// stored object. A storage failure cannot resurrect the deleted row, so it is
// logged and left to the orphan sweep.
type DeleteAttachmentUsecase struct {
	repo    attachmentout.AttachmentRepository
	storage attachmentout.FileStorage
	access  attachmentout.RefAccessChecker
	tx      attachmentout.TransactionManager
}

func NewDeleteAttachmentUsecase(
	repo attachmentout.AttachmentRepository,
	storage attachmentout.FileStorage,
	access attachmentout.RefAccessChecker,
	tx attachmentout.TransactionManager,
) *DeleteAttachmentUsecase {
	return &DeleteAttachmentUsecase{repo: repo, storage: storage, access: access, tx: tx}
}

func (uc *DeleteAttachmentUsecase) Execute(ctx context.Context, cmd attachmentin.DeleteAttachmentCommand) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	attachment, err := uc.repo.GetByID(ctx, cmd.AttachmentID)
	if err != nil {
		return attachmentRecordError(err, nil)
	}
	if err := attachmentRecordError(nil, attachment); err != nil {
		return err
	}

	ref := attachment.Ref()
	if err := uc.access.CanAccess(ctx, ref.RefName().String(), ref.ID(), cmd.ActorID, cmd.ActorRole, attachmentout.ActionDelete); err != nil {
		return err
	}

	var deferredToWorker bool
	if err := uc.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if attachment.Status() == domain.AttachmentStatusProcessing {
			deferredToWorker = true
			return uc.repo.UpdateStatus(txCtx, cmd.AttachmentID, domain.AttachmentStatusDeleting)
		}
		return uc.repo.Delete(txCtx, cmd.AttachmentID)
	}); err != nil {
		return err
	}
	if deferredToWorker {
		return nil
	}

	if err := uc.storage.Delete(ctx, attachment.StorageKey()); err != nil {
		slog.Warn("attachment object delete failed after metadata was removed",
			"attachment_id", attachment.ID(),
			"ref_name", ref.RefName().String(),
			"ref_id", ref.ID(),
			"error", sanitisedAttachmentError(err),
		)
	}

	return nil
}

package usecase

import (
	"context"
	"log/slog"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
)

var _ attachmentin.DeleteAttachmentsForRefUsecaseInt = (*DeleteAttachmentsForRefUsecase)(nil)

// DeleteAttachmentsForRefUsecase is called when a parent resource is removed.
// It is idempotent: no rows means no keys to clean up.
type DeleteAttachmentsForRefUsecase struct {
	repo    attachmentout.AttachmentRepository
	storage attachmentout.FileStorage
	tx      attachmentout.TransactionManager
}

func NewDeleteAttachmentsForRefUsecase(
	repo attachmentout.AttachmentRepository,
	storage attachmentout.FileStorage,
	tx attachmentout.TransactionManager,
) *DeleteAttachmentsForRefUsecase {
	return &DeleteAttachmentsForRefUsecase{repo: repo, storage: storage, tx: tx}
}

func (uc *DeleteAttachmentsForRefUsecase) Execute(ctx context.Context, cmd attachmentin.DeleteForRefCommand) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	ref, err := domain.NewAttachmentRef(domain.AttachmentRefName(cmd.RefName), cmd.RefID)
	if err != nil {
		return 0, err
	}
	refName, refID := ref.RefName().String(), ref.ID()

	var keys []string
	if err := uc.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		deletedKeys, err := uc.repo.DeleteByRef(txCtx, refName, refID)
		if err != nil {
			return err
		}
		keys = deletedKeys
		return nil
	}); err != nil {
		return 0, err
	}

	for _, key := range keys {
		if err := uc.storage.Delete(ctx, key); err != nil {
			slog.Warn("attachment object delete failed during ref cleanup",
				"ref_name", refName,
				"ref_id", refID,
				"error", sanitisedAttachmentError(err),
			)
		}
	}

	return len(keys), nil
}

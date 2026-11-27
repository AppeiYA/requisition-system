package usecase

import (
	"context"
	"log/slog"
	"time"

	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
)

var _ attachmentin.DownloadAttachmentUsecaseInt = (*DownloadAttachmentUsecase)(nil)

// DownloadAttachmentUsecase either opens an authenticated stream or hands out
// a short-lived signed URL, depending on whether the storage adapter can sign.
type DownloadAttachmentUsecase struct {
	repo         attachmentout.AttachmentRepository
	storage      attachmentout.FileStorage
	access       attachmentout.RefAccessChecker
	signer       attachmentout.SignedURLProvider
	signedURLTTL time.Duration
}

func NewDownloadAttachmentUsecase(
	repo attachmentout.AttachmentRepository,
	storage attachmentout.FileStorage,
	access attachmentout.RefAccessChecker,
	signer attachmentout.SignedURLProvider,
	signedURLTTL time.Duration,
) *DownloadAttachmentUsecase {
	return &DownloadAttachmentUsecase{
		repo:         repo,
		storage:      storage,
		access:       access,
		signer:       signer,
		signedURLTTL: signedURLTTL,
	}
}

func (uc *DownloadAttachmentUsecase) Execute(ctx context.Context, query attachmentin.DownloadAttachmentQuery) (*attachmentin.Download, error) {
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
	// A claimed upload has metadata but its bytes are still at staging/<upl-id>.
	// Never open them until reconciliation moves them and marks the row READY.
	if err := attachment.CanDownload(); err != nil {
		return nil, err
	}

	ref := attachment.Ref()
	if err := uc.access.CanAccess(ctx, ref.RefName().String(), ref.ID(), query.ActorID, query.ActorRole, attachmentout.ActionRead); err != nil {
		return nil, err
	}

	if uc.signer != nil {
		location, expiresAt, err := uc.signer.SignedURL(ctx, attachment.StorageKey(), uc.signedURLTTL, attachment.FileName().String())
		if err != nil {
			if attachmentout.IsObjectNotFound(err) {
				return nil, attachmentout.ErrObjectNotFound
			}
			return nil, err
		}
		return &attachmentin.Download{
			Attachment: attachment,
			Location:   location,
			ExpiresAt:  expiresAt,
		}, nil
	}

	stream, info, err := uc.storage.Open(ctx, attachment.StorageKey())
	if err != nil {
		if attachmentout.IsObjectNotFound(err) {
			return nil, attachmentout.ErrObjectNotFound
		}
		return nil, err
	}
	// Zero means the adapter did not report a size; only compare when the
	// adapter actually knows how many bytes it is handing back.
	if info.Size > 0 && info.Size != attachment.FileSize() {
		if stream != nil {
			_ = stream.Close()
		}
		slog.Warn("stored attachment size does not match metadata",
			"attachment_id", attachment.ID(),
			"ref_name", ref.RefName().String(),
			"ref_id", ref.ID(),
		)
		return nil, ErrAttachmentIntegrity
	}

	return &attachmentin.Download{
		Attachment: attachment,
		Stream:     stream,
	}, nil
}

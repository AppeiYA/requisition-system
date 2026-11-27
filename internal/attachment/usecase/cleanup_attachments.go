package usecase

import (
	"context"
	"log/slog"
	"time"

	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
)

var _ attachmentin.CleanupAttachmentsUsecaseInt = (*CleanupAttachmentsUsecase)(nil)

// CleanupAttachmentsUsecase expires overdue pending staged uploads (deleting
// their staged objects without touching claimed rows), recovers attachments stuck
// in PROCESSING longer than stuckAfter, and reports queue depth metrics.
type CleanupAttachmentsUsecase struct {
	stagedRepo     attachmentout.StagedUploadRepository
	attachmentRepo attachmentout.AttachmentRepository
	outboxRepo     attachmentout.ReconciliationOutboxRepository
	storage        attachmentout.FileStorage
	tx             attachmentout.TransactionManager
	metrics        attachmentout.MetricsRecorder
	batchSize      int
	stuckAfter     time.Duration
	now            func() time.Time
}

func NewCleanupAttachmentsUsecase(
	stagedRepo attachmentout.StagedUploadRepository,
	attachmentRepo attachmentout.AttachmentRepository,
	outboxRepo attachmentout.ReconciliationOutboxRepository,
	storage attachmentout.FileStorage,
	tx attachmentout.TransactionManager,
	metrics attachmentout.MetricsRecorder,
	batchSize int,
	stuckAfter time.Duration,
) *CleanupAttachmentsUsecase {
	if batchSize < 1 {
		batchSize = 50
	}
	if stuckAfter <= 0 {
		stuckAfter = 15 * time.Minute
	}
	return &CleanupAttachmentsUsecase{
		stagedRepo:     stagedRepo,
		attachmentRepo: attachmentRepo,
		outboxRepo:     outboxRepo,
		storage:        storage,
		tx:             tx,
		metrics:        metrics,
		batchSize:      batchSize,
		stuckAfter:     stuckAfter,
		now:            func() time.Time { return time.Now().UTC() },
	}
}

func (uc *CleanupAttachmentsUsecase) Execute(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	now := uc.now()

	if err := uc.expireStagedUploads(ctx, now); err != nil {
		return err
	}
	if err := uc.recoverStuckProcessing(ctx, now); err != nil {
		return err
	}
	uc.recordMetrics(ctx, now)
	return nil
}

func (uc *CleanupAttachmentsUsecase) expireStagedUploads(ctx context.Context, now time.Time) error {
	var expiredCount int64
	for {
		var batchKeys []string
		if err := uc.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
			expired, err := uc.stagedRepo.ExpirePending(txCtx, now, uc.batchSize)
			if err != nil {
				return err
			}
			for _, u := range expired {
				batchKeys = append(batchKeys, u.StoragePath())
			}
			return nil
		}); err != nil {
			return err
		}
		if len(batchKeys) == 0 {
			break
		}
		for _, key := range batchKeys {
			if err := uc.storage.Delete(ctx, key); err != nil {
				slog.WarnContext(ctx, "failed to delete expired staged upload object",
					"error", sanitisedAttachmentError(err),
				)
			} else {
				expiredCount++
			}
		}
		if len(batchKeys) < uc.batchSize {
			break
		}
	}
	if uc.metrics != nil && expiredCount > 0 {
		uc.metrics.AddCleanupDeletions(ctx, expiredCount)
	}
	return nil
}

func (uc *CleanupAttachmentsUsecase) recoverStuckProcessing(ctx context.Context, now time.Time) error {
	stuckBefore := now.Add(-uc.stuckAfter)
	return uc.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		ids, err := uc.attachmentRepo.ResetStuckProcessing(txCtx, stuckBefore, uc.batchSize)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		_, err = uc.outboxRepo.RequeueByAttachmentIDs(txCtx, ids, now)
		return err
	})
}

func (uc *CleanupAttachmentsUsecase) recordMetrics(ctx context.Context, now time.Time) {
	if uc.metrics == nil {
		return
	}
	stagedPending, err := uc.stagedRepo.CountPendingAll(ctx, now)
	if err != nil {
		return
	}
	outboxDepth, deadEvents, err := uc.outboxRepo.CountDepth(ctx)
	if err != nil {
		return
	}
	uc.metrics.RecordQueueStats(ctx, stagedPending, outboxDepth, deadEvents)
}

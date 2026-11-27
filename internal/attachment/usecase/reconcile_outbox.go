package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
)

var _ attachmentin.ReconcileOutboxUsecaseInt = (*ReconcileOutboxUsecase)(nil)

// ReconcileOutboxUsecase moves claimed files to their permanent keys.
//
// One event is claimed, moved and finished inside a single transaction. A
// crash therefore rolls the claim back to pending/failed instead of stranding
// an event as processing. FileStorage.Move is idempotent, so a commit failure
// after a successful move is safe to retry.
type ReconcileOutboxUsecase struct {
	outbox      attachmentout.ReconciliationOutboxRepository
	attachments attachmentout.AttachmentRepository
	storage     attachmentout.FileStorage
	tx          attachmentout.TransactionManager
	metrics     attachmentout.MetricsRecorder
	batchSize   int
	maxAttempts int
	backoff     time.Duration
	now         func() time.Time
}

func NewReconcileOutboxUsecase(
	outbox attachmentout.ReconciliationOutboxRepository,
	attachments attachmentout.AttachmentRepository,
	storage attachmentout.FileStorage,
	tx attachmentout.TransactionManager,
	metrics attachmentout.MetricsRecorder,
	batchSize int,
	maxAttempts int,
	backoff time.Duration,
) *ReconcileOutboxUsecase {
	if batchSize < 1 {
		batchSize = 1
	}
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	if backoff <= 0 {
		backoff = time.Second
	}
	return &ReconcileOutboxUsecase{
		outbox: outbox, attachments: attachments, storage: storage, tx: tx, metrics: metrics,
		batchSize: batchSize, maxAttempts: maxAttempts, backoff: backoff,
		now: func() time.Time { return time.Now().UTC() },
	}
}

// Execute drains at most batchSize due events. No work left is not an error:
// an empty sweep is the normal outcome between requisition creations.
func (uc *ReconcileOutboxUsecase) Execute(ctx context.Context) error {
	for index := 0; index < uc.batchSize; index++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		drained, err := uc.reconcileOne(ctx)
		if err != nil {
			return err
		}
		if drained {
			return nil
		}
	}
	return nil
}

func (uc *ReconcileOutboxUsecase) reconcileOne(ctx context.Context) (bool, error) {
	drained := false
	err := uc.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		now := uc.now()
		event, err := uc.outbox.ClaimDue(txCtx, now)
		if err != nil {
			if errors.Is(err, attachmentout.ErrNoDueEvents) {
				drained = true
				return nil
			}
			return err
		}
		return uc.processEvent(txCtx, event, now)
	})
	return drained, err
}

func (uc *ReconcileOutboxUsecase) processEvent(
	ctx context.Context,
	event *domain.ReconciliationEvent,
	now time.Time,
) error {
	attachment, err := uc.attachments.GetByID(ctx, event.AttachmentID())
	if err != nil {
		if errors.Is(attachmentRecordError(err, nil), ErrAttachmentNotFound) {
			// A metadata row cannot disappear while its outbox row exists in a
			// healthy database. There is nothing to move, so make this event
			// terminal instead of retrying forever.
			return uc.deadLetter(ctx, event, now)
		}
		return err
	}

	if attachment.Status() == domain.AttachmentStatusDeleting {
		_ = uc.storage.Delete(ctx, event.ToKey())
		_ = uc.storage.Delete(ctx, event.FromKey())
		if delErr := uc.attachments.Delete(ctx, attachment.ID()); delErr != nil &&
			!errors.Is(attachmentRecordError(delErr, nil), ErrAttachmentNotFound) {
			return delErr
		}
		return uc.outbox.MarkDone(ctx, event.ID(), now)
	}
	if attachment.Status() == domain.AttachmentStatusReady {
		// Move already committed on an earlier attempt whose status update
		// failed. Treat the durable READY row as the source of truth.
		return uc.outbox.MarkDone(ctx, event.ID(), now)
	}
	if attachment.Status() == domain.AttachmentStatusFailed {
		return uc.deadLetter(ctx, event, now)
	}
	if attachment.Status() == domain.AttachmentStatusPending {
		if err := attachment.MarkProcessing(); err != nil {
			return err
		}
		if err := uc.attachments.UpdateStatus(ctx, attachment.ID(), attachment.Status()); err != nil {
			if errors.Is(attachmentRecordError(err, nil), ErrAttachmentNotFound) {
				_ = uc.storage.Delete(ctx, event.FromKey())
				return uc.outbox.MarkDone(ctx, event.ID(), now)
			}
			return err
		}
	}

	info, err := uc.storage.Move(ctx, event.FromKey(), event.ToKey(), event.Checksum())
	if err != nil {
		if attachmentout.IsObjectNotFound(err) {
			return uc.failTerminal(ctx, event, attachment, now, err)
		}
		return uc.fail(ctx, event, attachment, now, err)
	}
	if (info.Size > 0 && info.Size != attachment.FileSize()) ||
		(info.Checksum != "" && info.Checksum != attachment.Checksum()) {
		return uc.fail(ctx, event, attachment, now, attachmentout.ErrChecksumMismatch)
	}
	if err := attachment.MarkReady(); err != nil {
		return uc.fail(ctx, event, attachment, now, err)
	}
	if err := uc.attachments.UpdateStatus(ctx, attachment.ID(), attachment.Status()); err != nil {
		if errors.Is(attachmentRecordError(err, nil), ErrAttachmentNotFound) {
			// Attachment was deleted while we were moving the file; clean up
			// the moved object so no orphan remains and complete the event.
			_ = uc.storage.Delete(ctx, event.ToKey())
			return uc.outbox.MarkDone(ctx, event.ID(), now)
		}
		return err
	}
	return uc.outbox.MarkDone(ctx, event.ID(), now)
}

func (uc *ReconcileOutboxUsecase) failTerminal(
	ctx context.Context,
	event *domain.ReconciliationEvent,
	attachment *domain.Attachment,
	now time.Time,
	cause error,
) error {
	if attachment != nil {
		if err := attachment.MarkFailed(); err == nil {
			if err := uc.attachments.UpdateStatus(ctx, attachment.ID(), attachment.Status()); err != nil &&
				!errors.Is(attachmentRecordError(err, nil), ErrAttachmentNotFound) {
				return err
			}
		}
	}
	slog.WarnContext(ctx, "attachment reconciliation permanently failed",
		slog.String("event.id", event.ID()),
		slog.String("attachment.id", event.AttachmentID()),
		slog.String("event.status", domain.ReconciliationStatusDead.String()),
		slog.Int("event.attempts", event.Attempts()+1),
		slog.String("error.message", cause.Error()),
	)
	return uc.deadLetter(ctx, event, now)
}

// deadLetter records a terminal outbox transition and reports it once, after
// the durable write succeeds. Callers are responsible for logging the cause.
func (uc *ReconcileOutboxUsecase) deadLetter(
	ctx context.Context,
	event *domain.ReconciliationEvent,
	now time.Time,
) error {
	err := uc.outbox.RecordFailure(
		ctx, event.ID(), domain.ReconciliationStatusDead,
		event.Attempts()+1, now, now,
	)
	if err != nil {
		return err
	}
	if uc.metrics != nil {
		uc.metrics.AddDeadLettered(ctx, 1)
	}
	return nil
}

// fail records a retry or a terminal failure without rolling back the claim:
// the event leaves PROCESSING with its attempt count and next-attempt time
// updated, and the attachment remains non-downloadable either way.
func (uc *ReconcileOutboxUsecase) fail(
	ctx context.Context,
	event *domain.ReconciliationEvent,
	attachment *domain.Attachment,
	now time.Time,
	cause error,
) error {
	status, attempts, nextAttemptAt := event.FailureSchedule(uc.maxAttempts, uc.backoff, now)

	if status == domain.ReconciliationStatusDead && attachment != nil {
		if err := attachment.MarkFailed(); err == nil {
			if err := uc.attachments.UpdateStatus(ctx, attachment.ID(), attachment.Status()); err != nil {
				return err
			}
		}
	}

	slog.WarnContext(ctx, "attachment reconciliation attempt failed",
		slog.String("event.id", event.ID()),
		slog.String("attachment.id", event.AttachmentID()),
		slog.String("event.status", status.String()),
		slog.Int("event.attempts", attempts),
		slog.String("error.message", cause.Error()),
	)

	if status == domain.ReconciliationStatusDead {
		return uc.deadLetter(ctx, event, now)
	}
	return uc.outbox.RecordFailure(ctx, event.ID(), status, attempts, nextAttemptAt, now)
}

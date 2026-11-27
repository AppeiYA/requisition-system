package attachmentout

import (
	"context"
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrReconciliationEventExists   = customerrors.NewConflictError("attachment reconciliation event already exists")
	ErrReconciliationEventNotFound = customerrors.NewNotFoundError("attachment reconciliation event not found")
	// ErrNoDueEvents is the normal stop signal for a reconciliation sweep.
	ErrNoDueEvents = customerrors.NewNotFoundError("no reconciliation events are due")
	// ErrReconciliationEventConflict means a status guard failed because
	// another worker changed the row first.
	ErrReconciliationEventConflict = customerrors.NewConflictError("reconciliation event state changed")
)

// ReconciliationOutboxRepository writes durable move instructions in the same
// transaction that creates the requisition, and lets the worker claim and
// finish them. ClaimDue must be called inside a transaction: the claim, the
// storage move and the final status all commit or roll back together.
type ReconciliationOutboxRepository interface {
	Save(ctx context.Context, event *domain.ReconciliationEvent) error
	GetByID(ctx context.Context, id string) (*domain.ReconciliationEvent, error)
	GetByAttachmentID(ctx context.Context, attachmentID string) (*domain.ReconciliationEvent, error)

	// ClaimDue atomically takes one due pending/failed event. The caller's
	// transaction still owns the row when this returns, so a crash rolls the
	// claim back to its previous state instead of stranding it as processing.
	ClaimDue(ctx context.Context, now time.Time) (*domain.ReconciliationEvent, error)

	// MarkDone and RecordFailure require a row this worker currently owns.
	MarkDone(ctx context.Context, id string, now time.Time) error
	RecordFailure(
		ctx context.Context,
		id string,
		status domain.ReconciliationStatus,
		attempts int,
		nextAttemptAt time.Time,
		now time.Time,
	) error
	RequeueByAttachmentIDs(ctx context.Context, attachmentIDs []string, now time.Time) (int64, error)
	CountDepth(ctx context.Context) (pendingOrFailed int64, dead int64, err error)
}

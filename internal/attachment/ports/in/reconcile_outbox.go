package attachmentin

import "context"

// ReconcileOutboxUsecaseInt processes a bounded batch of due attachment
// reconciliation events. It is driven by the worker, never by an HTTP request.
type ReconcileOutboxUsecaseInt interface {
	Execute(ctx context.Context) error
}

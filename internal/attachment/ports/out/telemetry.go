package attachmentout

import "context"

const (
	OperationUpload    = "attachment.upload"
	OperationDownload  = "attachment.download"
	OperationDelete    = "attachment.delete"
	OperationReconcile = "attachment.reconcile"
	OperationCleanup   = "attachment.cleanup"
)

// MetricsRecorder reports low-cardinality attachment queue and cleanup metrics.
// Implementations must never attach IDs, keys, or filenames as attributes.
type MetricsRecorder interface {
	RecordQueueStats(ctx context.Context, stagedPending, outboxDepth, deadEvents int64)
	AddCleanupDeletions(ctx context.Context, count int64)
	AddDeadLettered(ctx context.Context, count int64)
}

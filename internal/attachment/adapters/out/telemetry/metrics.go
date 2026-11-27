package attachmenttelemetry

import (
	"context"

	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/telemetry"
)

var _ attachmentout.MetricsRecorder = (*Recorder)(nil)

// Recorder publishes low-cardinality attachment queue and cleanup metrics.
// No IDs, keys, or filenames are ever attached as attributes.
type Recorder struct {
	stagedPending     *telemetry.Gauge
	outboxDepth       *telemetry.Gauge
	deadEvents        *telemetry.Gauge
	cleanupDeletions  *telemetry.Counter
	deadLetteredTotal *telemetry.Counter
}

func NewRecorder() (*Recorder, error) {
	stagedPending, err := telemetry.NewGauge(
		"requisition.attachment.staged_pending",
		"Number of pending unexpired staged uploads",
		"{upload}",
	)
	if err != nil {
		return nil, err
	}
	outboxDepth, err := telemetry.NewGauge(
		"requisition.attachment.outbox_depth",
		"Number of active attachment reconciliation outbox events",
		"{event}",
	)
	if err != nil {
		return nil, err
	}
	deadEvents, err := telemetry.NewGauge(
		"requisition.attachment.dead_lettered_events",
		"Number of dead-lettered attachment reconciliation outbox events",
		"{event}",
	)
	if err != nil {
		return nil, err
	}
	cleanupDeletions, err := telemetry.NewCounter(
		"requisition.attachment.cleanup_deletions",
		"Total expired staged upload objects removed by cleanup",
		"{upload}",
	)
	if err != nil {
		return nil, err
	}
	deadLetteredTotal, err := telemetry.NewCounter(
		"requisition.attachment.dead_lettered_total",
		"Total attachment reconciliation events moved to dead-letter",
		"{event}",
	)
	if err != nil {
		return nil, err
	}

	return &Recorder{
		stagedPending:     stagedPending,
		outboxDepth:       outboxDepth,
		deadEvents:        deadEvents,
		cleanupDeletions:  cleanupDeletions,
		deadLetteredTotal: deadLetteredTotal,
	}, nil
}

func (r *Recorder) RecordQueueStats(ctx context.Context, stagedPending, outboxDepth, deadEvents int64) {
	r.stagedPending.Record(ctx, stagedPending)
	r.outboxDepth.Record(ctx, outboxDepth)
	r.deadEvents.Record(ctx, deadEvents)
}

func (r *Recorder) AddCleanupDeletions(ctx context.Context, count int64) {
	if count > 0 {
		r.cleanupDeletions.Add(ctx, count)
	}
}

func (r *Recorder) AddDeadLettered(ctx context.Context, count int64) {
	if count > 0 {
		r.deadLetteredTotal.Add(ctx, count)
	}
}

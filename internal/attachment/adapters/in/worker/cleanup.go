package attachmentworker

import (
	"context"
	"log/slog"

	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/telemetry"
	"github.com/hibiken/asynq"
)

// Cleanup is the scheduled task that expires overdue staged uploads and
// recovers attachments stuck in PROCESSING.
const Cleanup = "attachment:cleanup"

type CleanupAttachmentsHandler struct {
	cleanup   attachmentin.CleanupAttachmentsUsecaseInt
	telemetry *telemetry.OperationTelemetry
}

func NewCleanupAttachmentsHandler(
	cleanup attachmentin.CleanupAttachmentsUsecaseInt,
	operationTelemetry *telemetry.OperationTelemetry,
) *CleanupAttachmentsHandler {
	return &CleanupAttachmentsHandler{cleanup: cleanup, telemetry: operationTelemetry}
}

func (h *CleanupAttachmentsHandler) Handle(ctx context.Context, _ *asynq.Task) error {
	run := func(ctx context.Context) error {
		if err := h.cleanup.Execute(ctx); err != nil {
			slog.ErrorContext(ctx, "attachment cleanup sweep failed",
				slog.String("error.message", err.Error()),
			)
			return err
		}
		return nil
	}

	if h.telemetry == nil {
		return run(ctx)
	}
	return h.telemetry.Measure(ctx, attachmentout.OperationCleanup, run)
}

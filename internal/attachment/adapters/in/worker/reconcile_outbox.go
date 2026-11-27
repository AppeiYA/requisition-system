package attachmentworker

import (
	"context"
	"log/slog"

	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/telemetry"
	"github.com/hibiken/asynq"
)

// ReconcileOutbox is the scheduled task that drains due attachment moves.
const ReconcileOutbox = "attachment:reconcile"

// ReconcileOutboxHandler is the inbound adapter from asynq to the attachment
// reconciliation use case. It owns no storage or repository logic.
type ReconcileOutboxHandler struct {
	reconcile attachmentin.ReconcileOutboxUsecaseInt
	telemetry *telemetry.OperationTelemetry
}

func NewReconcileOutboxHandler(
	reconcile attachmentin.ReconcileOutboxUsecaseInt,
	operationTelemetry *telemetry.OperationTelemetry,
) *ReconcileOutboxHandler {
	return &ReconcileOutboxHandler{reconcile: reconcile, telemetry: operationTelemetry}
}

func (h *ReconcileOutboxHandler) Handle(ctx context.Context, _ *asynq.Task) error {
	run := func(ctx context.Context) error {
		if err := h.reconcile.Execute(ctx); err != nil {
			slog.ErrorContext(ctx, "attachment reconciliation sweep failed",
				slog.String("error.message", err.Error()),
			)
			return err
		}
		return nil
	}

	if h.telemetry == nil {
		return run(ctx)
	}
	return h.telemetry.Measure(ctx, attachmentout.OperationReconcile, run)
}

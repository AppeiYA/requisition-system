package identityworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	identitytasks "github.com/AppeiYA/requisition-system/internal/identity/tasks"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/telemetry"
	"github.com/hibiken/asynq"
)

// PasswordResetQueue hands reset emails to asynq. The API enqueues through
// it; the worker process drains the task.
type PasswordResetQueue struct {
	client *asynq.Client
	cfg    config.WorkerConfig
}

func NewPasswordResetQueue(client *asynq.Client, cfg config.WorkerConfig) *PasswordResetQueue {
	return &PasswordResetQueue{client: client, cfg: cfg}
}

var _ identityout.PasswordResetEmailQueue = (*PasswordResetQueue)(nil)

func (q *PasswordResetQueue) EnqueuePasswordResetEmail(ctx context.Context, payload identitytasks.PasswordResetEmailPayload) error {
	payload.TraceParent = telemetry.TraceParentFrom(ctx)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s payload: %w", identitytasks.SendPasswordResetEmail, err)
	}
	task := asynq.NewTask(identitytasks.SendPasswordResetEmail, body)

	// The task ID is derived from the reset ID, never from the raw token.
	_, err = q.client.EnqueueContext(ctx, task,
		asynq.Queue(q.cfg.Queue),
		asynq.MaxRetry(q.cfg.TaskMaxRetry),
		asynq.Timeout(q.cfg.TaskTimeout),
		asynq.TaskID("password-reset-email:"+payload.ResetID),
	)
	if errors.Is(err, asynq.ErrTaskIDConflict) || errors.Is(err, asynq.ErrDuplicateTask) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("enqueue %s task: %w", identitytasks.SendPasswordResetEmail, err)
	}
	return nil
}

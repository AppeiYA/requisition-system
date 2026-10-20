package adminworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	"github.com/AppeiYA/requisition-system/internal/admin/tasks"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/hibiken/asynq"
)

// TaskQueue is the outbound adapter that hands invitation emails to asynq.
// The API process enqueues through it; the worker process drains it.
type TaskQueue struct {
	client *asynq.Client
	cfg    config.WorkerConfig
}

func NewTaskQueue(client *asynq.Client, cfg config.WorkerConfig) *TaskQueue {
	return &TaskQueue{
		client: client,
		cfg:    cfg,
	}
}

var _ adminout.TaskQueue = (*TaskQueue)(nil)

func (q *TaskQueue) EnqueueInvitationEmail(ctx context.Context, payload tasks.InvitationEmailPayload) error {
	task, err := invitationEmailTask(payload)
	if err != nil {
		return err
	}

	// The task ID derives from the invitation token, so retrying the API
	// request can never queue the same invitation mail twice. A conflict
	// means the task is already queued and is treated as success.
	_, err = q.client.EnqueueContext(ctx, task,
		asynq.Queue(q.cfg.Queue),
		asynq.MaxRetry(q.cfg.TaskMaxRetry),
		asynq.Timeout(q.cfg.TaskTimeout),
		asynq.TaskID("invitation-email:"+payload.Token),
	)
	if errors.Is(err, asynq.ErrTaskIDConflict) || errors.Is(err, asynq.ErrDuplicateTask) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("enqueue %s task: %w", tasks.SendInvitationEmail, err)
	}
	return nil
}

// invitationEmailTask renders the task contract shared with the worker-side
// handler. It is kept separate so the encoding can be tested without Redis.
func invitationEmailTask(payload tasks.InvitationEmailPayload) (*asynq.Task, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal %s payload: %w", tasks.SendInvitationEmail, err)
	}
	return asynq.NewTask(tasks.SendInvitationEmail, b), nil
}

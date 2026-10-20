package adminworker

import (
	"context"
	"encoding/json"
	"fmt"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	"github.com/AppeiYA/requisition-system/internal/admin/tasks"
	"github.com/hibiken/asynq"
)

// SendMailHandler is the inbound adapter that receives queued invitation
// email tasks from asynq and drives the application through its in-port,
// the same way the HTTP handlers do.
type SendMailHandler struct {
	sendInvitationEmail adminin.SendInvitationEmailUseCase
}

func NewSendMailHandler(sendInvitationEmail adminin.SendInvitationEmailUseCase) *SendMailHandler {
	return &SendMailHandler{
		sendInvitationEmail: sendInvitationEmail,
	}
}

// Handle processes a single task. A payload that cannot be decoded will
// never decode on a retry, so it skips retries; delivery failures are
// returned and retried by asynq with exponential backoff.
func (h *SendMailHandler) Handle(ctx context.Context, t *asynq.Task) error {
	var payload tasks.InvitationEmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("%w: decode %s payload: %v", asynq.SkipRetry, t.Type(), err)
	}

	if err := h.sendInvitationEmail.Execute(ctx, adminin.SendInvitationEmailCommand{
		To:             payload.To,
		InvitedBy:      payload.InvitedBy,
		Token:          payload.Token,
		ExpiresAt:      payload.ExpiresAt,
		DepartmentName: payload.DepartmentName,
		JobTitle:       payload.JobTitle,
	}); err != nil {
		return fmt.Errorf("send invitation email: %w", err)
	}

	return nil
}

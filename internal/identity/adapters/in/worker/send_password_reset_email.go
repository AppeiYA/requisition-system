package identityworkerin

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identitytasks "github.com/AppeiYA/requisition-system/internal/identity/tasks"
	"github.com/hibiken/asynq"
)

// SendPasswordResetEmailHandler receives queued reset emails. A payload that
// cannot be decoded will never decode on a retry, so it skips retries; a
// delivery failure is returned so asynq retries it with backoff.
type SendPasswordResetEmailHandler struct {
	usecase identityin.SendPasswordResetEmailUsecaseInt
}

func NewSendPasswordResetEmailHandler(usecase identityin.SendPasswordResetEmailUsecaseInt) *SendPasswordResetEmailHandler {
	return &SendPasswordResetEmailHandler{usecase: usecase}
}

func (h *SendPasswordResetEmailHandler) Handle(ctx context.Context, task *asynq.Task) error {
	var payload identitytasks.PasswordResetEmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("%w: decode %s payload: %v", asynq.SkipRetry, task.Type(), err)
	}
	if payload.To == "" || payload.Token == "" {
		return fmt.Errorf("%w: incomplete %s payload", asynq.SkipRetry, task.Type())
	}
	expiresAt, err := time.Parse(time.RFC3339, payload.ExpiresAt)
	if err != nil {
		return fmt.Errorf("%w: invalid expiry in %s payload: %v", asynq.SkipRetry, task.Type(), err)
	}

	if err := h.usecase.Execute(ctx, identityin.SendPasswordResetEmailCommand{
		To:        payload.To,
		Token:     payload.Token,
		ExpiresAt: expiresAt,
	}); err != nil {
		// Never log the recipient or the token: only the reset correlation ID.
		slog.ErrorContext(ctx, "password_reset.email_failed",
			slog.String("event", "password_reset.email_failed"),
			slog.String("reset.id", payload.ResetID),
			slog.String("reason", "delivery_failed"),
		)
		return fmt.Errorf("send password reset email: %w", err)
	}

	slog.InfoContext(ctx, "password_reset.email_sent",
		slog.String("event", "password_reset.email_sent"),
		slog.String("reset.id", payload.ResetID),
	)
	return nil
}

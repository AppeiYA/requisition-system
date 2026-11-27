package identityin

import (
	"context"
	"time"
)

// RequestPasswordResetUsecaseInt starts the self-service reset flow. It always
// returns nil for an unknown, inactive, or throttled account so the HTTP layer
// can answer with one generic message.
type RequestPasswordResetUsecaseInt interface {
	Execute(ctx context.Context, email string) error
}

// ResetPasswordUsecaseInt redeems a reset token and sets a new password.
type ResetPasswordUsecaseInt interface {
	Execute(ctx context.Context, token, password string) error
}

// SendPasswordResetEmailCommand carries everything needed to deliver one reset
// email. It is driven by the worker process through the queued task.
type SendPasswordResetEmailCommand struct {
	To        string
	Token     string
	ExpiresAt time.Time
}

// SendPasswordResetEmailUsecaseInt delivers a reset email from the worker.
type SendPasswordResetEmailUsecaseInt interface {
	Execute(ctx context.Context, cmd SendPasswordResetEmailCommand) error
}

// CleanupPasswordResetTokensUsecaseInt removes terminal rows past retention.
type CleanupPasswordResetTokensUsecaseInt interface {
	Execute(ctx context.Context) (int64, error)
}

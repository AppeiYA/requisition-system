package identityout

import (
	"context"
	"time"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identitytasks "github.com/AppeiYA/requisition-system/internal/identity/tasks"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

// ErrPasswordResetTokenNotFound is returned by the repository when a token is
// unknown, expired, already terminal, or belongs to another user. The use cases
// translate it to a generic 400 before it reaches HTTP.
var ErrPasswordResetTokenNotFound = customerrors.NewNotFoundError("password reset token not found")

// PasswordResetTokenRepository persists password reset tokens. Every method
// uses the executor carried by ctx, so it participates in the ambient
// transaction when one is present.
type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *domain.PasswordResetToken) error

	// GetByTokenHash is a read-only pre-check used to learn the user ID before
	// taking the user row lock. Consume is the authoritative check.
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error)

	// GetLatestActiveForUser returns the newest token with used_at IS NULL and
	// expires_at > now, or (nil, nil) when there is none.
	GetLatestActiveForUser(ctx context.Context, userID string, now time.Time) (*domain.PasswordResetToken, error)

	// CountCreatedSince supports the per-account hourly cap.
	CountCreatedSince(ctx context.Context, userID string, since time.Time) (int, error)

	// InvalidateActiveForUser terminally marks every redeemable token for the
	// user as used. It is idempotent and is only called while holding the user
	// row lock.
	InvalidateActiveForUser(ctx context.Context, userID string, at time.Time) error

	// InvalidateByID terminally marks one specific token. The request flow uses
	// it after a failed enqueue so it cannot accidentally invalidate a newer
	// token created by a concurrent request.
	InvalidateByID(ctx context.Context, id string, at time.Time) error

	// Consume atomically marks one token terminal and returns it. It must
	// return ErrPasswordResetTokenNotFound when the token is unknown, expired,
	// already terminal, or belongs to another user.
	Consume(ctx context.Context, tokenHash, userID string, now time.Time) (*domain.PasswordResetToken, error)

	// DeleteOlderThan removes rows created before the cutoff.
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}

// PasswordResetEmailQueue hands reset emails to the background worker.
type PasswordResetEmailQueue interface {
	EnqueuePasswordResetEmail(ctx context.Context, payload identitytasks.PasswordResetEmailPayload) error
}

// PasswordResetEmailSender delivers a reset link through the configured email
// provider.
type PasswordResetEmailSender interface {
	SendPasswordResetEmail(ctx context.Context, to, token string, expiresAt time.Time) error
}

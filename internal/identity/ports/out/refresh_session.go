package identityout

import (
	"context"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

// Application-level errors returned by RefreshSessionRepository. They are
// *customerrors.CustomError values so the HTTP adapter can map them onto the
// right status code and callers can match them with errors.Is.
var (
	// ErrRefreshSessionNotFound is returned when no active session exists for
	// the presented token (unknown, already consumed or expired).
	ErrRefreshSessionNotFound = customerrors.NewUnauthorizedError("refresh session not found")

	// ErrRefreshSessionRevoked is returned when the token's session family has
	// been revoked, e.g. by logout or by reuse detection.
	ErrRefreshSessionRevoked = customerrors.NewUnauthorizedError("refresh session revoked")

	// ErrRefreshSessionReused is returned when an already-consumed token is
	// presented again. The implementation revokes the whole family before
	// returning this error.
	ErrRefreshSessionReused = customerrors.NewUnauthorizedError("refresh token reuse detected")

	// ErrRefreshSessionStore is returned when the session store cannot be
	// reached or answers unexpectedly. It maps to HTTP 500 and hides
	// infrastructure detail from callers.
	ErrRefreshSessionStore = customerrors.NewInternalServerError("refresh session store unavailable")
)

// RefreshSession is the application-level view of a persisted refresh session.
// It deliberately carries no storage-specific detail.
type RefreshSession struct {
	UserID   string
	FamilyID string
}

// RefreshSessionRepository persists refresh sessions. Implementations must
// treat Rotate as an atomic, single-use operation.
type RefreshSessionRepository interface {
	// Create starts a new session family for userID and stores the session
	// identified by tokenHash.
	Create(ctx context.Context, userID, tokenHash string) error

	// Rotate atomically consumes the session for oldTokenHash and stores a
	// replacement for newTokenHash in the same family, returning the consumed
	// session. Two concurrent calls with the same oldTokenHash must never both
	// succeed.
	Rotate(ctx context.Context, oldTokenHash, newTokenHash string) (*RefreshSession, error)

	// Revoke deletes the session for tokenHash and revokes its whole family.
	// It is idempotent.
	Revoke(ctx context.Context, tokenHash string) error
}

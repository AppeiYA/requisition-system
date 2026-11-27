package domain

import (
	"time"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

const (
	// PasswordResetTokenBytes is the entropy of a generated reset token.
	// 32 bytes (256 bits) makes the token infeasible to guess.
	PasswordResetTokenBytes = 32

	// DefaultPasswordResetTokenTTL is how long a reset link stays valid when
	// the deployment does not override it.
	DefaultPasswordResetTokenTTL = 30 * time.Minute

	// PasswordResetTokenIDPrefix is the prefix of a reset row identifier.
	PasswordResetTokenIDPrefix = "prt"
)

var (
	ErrPasswordResetTokenIDRequired   = customerrors.NewInternalServerError("password reset token ID is required")
	ErrPasswordResetTokenUserRequired = customerrors.NewInternalServerError("password reset token user ID is required")
	ErrPasswordResetTokenHashRequired = customerrors.NewInternalServerError("password reset token hash is required")
	ErrPasswordResetTokenTTLInvalid   = customerrors.NewInternalServerError("password reset token TTL must be positive")
	ErrPasswordResetTokenAlreadyUsed  = customerrors.NewBadRequestError("password reset token has already been used")
)

// PasswordResetToken is the persisted state of one reset request. The raw
// token is never part of the entity: only its storage hash is kept.
type PasswordResetToken struct {
	id        string
	userID    string
	tokenHash string
	expiresAt time.Time
	usedAt    *time.Time
	createdAt time.Time
}

// NewPasswordResetToken builds a redeemable token that expires ttl after now.
func NewPasswordResetToken(id, userID, tokenHash string, now time.Time, ttl time.Duration) (*PasswordResetToken, error) {
	if id == "" {
		return nil, ErrPasswordResetTokenIDRequired
	}
	if userID == "" {
		return nil, ErrPasswordResetTokenUserRequired
	}
	if tokenHash == "" {
		return nil, ErrPasswordResetTokenHashRequired
	}
	if ttl <= 0 {
		return nil, ErrPasswordResetTokenTTLInvalid
	}

	now = now.UTC()

	return &PasswordResetToken{
		id:        id,
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: now.Add(ttl),
		createdAt: now,
	}, nil
}

// ReconstitutePasswordResetToken rebuilds the entity from persisted columns.
func ReconstitutePasswordResetToken(
	id,
	userID,
	tokenHash string,
	expiresAt time.Time,
	usedAt *time.Time,
	createdAt time.Time,
) *PasswordResetToken {
	return &PasswordResetToken{
		id:        id,
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		usedAt:    usedAt,
		createdAt: createdAt,
	}
}

func (t *PasswordResetToken) ID() string {
	return t.id
}

func (t *PasswordResetToken) UserID() string {
	return t.userID
}

func (t *PasswordResetToken) TokenHash() string {
	return t.tokenHash
}

func (t *PasswordResetToken) ExpiresAt() time.Time {
	return t.expiresAt
}

func (t *PasswordResetToken) UsedAt() *time.Time {
	return t.usedAt
}

func (t *PasswordResetToken) CreatedAt() time.Time {
	return t.createdAt
}

// IsExpired reports whether the token can no longer be redeemed at now.
func (t *PasswordResetToken) IsExpired(now time.Time) bool {
	return !t.expiresAt.After(now)
}

// IsUsed reports whether the token has reached a terminal state, either
// because it was redeemed or because it was superseded or invalidated.
func (t *PasswordResetToken) IsUsed() bool {
	return t.usedAt != nil
}

// IsActive reports whether the token can still be redeemed at now.
func (t *PasswordResetToken) IsActive(now time.Time) bool {
	return !t.IsUsed() && !t.IsExpired(now)
}

// MarkUsed moves the token to its terminal state. It rejects a token that is
// already terminal so a caller cannot silently overwrite the first use.
func (t *PasswordResetToken) MarkUsed(at time.Time) error {
	if t.usedAt != nil {
		return ErrPasswordResetTokenAlreadyUsed
	}
	usedAt := at.UTC()
	t.usedAt = &usedAt
	return nil
}

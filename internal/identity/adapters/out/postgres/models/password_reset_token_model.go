package models

import (
	"time"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

// PasswordResetTokenModel maps one password_reset_tokens row. The token_hash
// column holds the SHA-256 hex of the raw token; the raw token is never stored.
type PasswordResetTokenModel struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
	CreatedAt time.Time  `db:"created_at"`
}

func (m PasswordResetTokenModel) ToDomain() *domain.PasswordResetToken {
	return domain.ReconstitutePasswordResetToken(
		m.ID,
		m.UserID,
		m.TokenHash,
		m.ExpiresAt,
		m.UsedAt,
		m.CreatedAt,
	)
}

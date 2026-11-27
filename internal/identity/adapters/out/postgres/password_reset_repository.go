package identitypostgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres/models"
	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
)

const passwordResetTokenColumns = `
	id, user_id, token_hash, expires_at, used_at, created_at
`

var (
	createPasswordResetTokenQuery = `
		INSERT INTO password_reset_tokens
			(id, user_id, token_hash, expires_at, created_at)
		VALUES (:id, :user_id, :token_hash, :expires_at, :created_at)
	`

	getPasswordResetTokenByHashQuery = `
		SELECT` + passwordResetTokenColumns + `
		FROM password_reset_tokens
		WHERE token_hash = $1
	`

	getLatestActivePasswordResetTokenQuery = `
		SELECT` + passwordResetTokenColumns + `
		FROM password_reset_tokens
		WHERE user_id = $1 AND used_at IS NULL AND expires_at > $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	countPasswordResetTokensQuery = `
		SELECT COUNT(*) FROM password_reset_tokens
		WHERE user_id = $1 AND created_at >= $2
	`

	invalidatePasswordResetTokensForUserQuery = `
		UPDATE password_reset_tokens
		SET used_at = $2
		WHERE user_id = $1 AND used_at IS NULL
	`

	invalidatePasswordResetTokenByIDQuery = `
		UPDATE password_reset_tokens
		SET used_at = $2
		WHERE id = $1 AND used_at IS NULL
	`

	consumePasswordResetTokenQuery = `
		UPDATE password_reset_tokens
		SET used_at = $3
		WHERE token_hash = $1
		  AND user_id = $2
		  AND used_at IS NULL
		  AND expires_at > $3
		RETURNING` + passwordResetTokenColumns + `
	`

	deletePasswordResetTokensQuery = `
		DELETE FROM password_reset_tokens WHERE created_at < $1
	`
)

// PasswordResetTokenRepository stores reset tokens with the same repository
// and executor as the user repository, so an ambient transaction is honoured
// by every method.
type PasswordResetTokenRepository struct {
	repository db.Repository
}

func NewPasswordResetTokenRepository(repository db.Repository) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{repository: repository}
}

var _ identityout.PasswordResetTokenRepository = (*PasswordResetTokenRepository)(nil)

func (r *PasswordResetTokenRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	_, err := r.repository.Executor(ctx).NamedExecContext(ctx, createPasswordResetTokenQuery, map[string]any{
		"id":         token.ID(),
		"user_id":    token.UserID(),
		"token_hash": token.TokenHash(),
		"expires_at": token.ExpiresAt(),
		"created_at": token.CreatedAt(),
	})
	return err
}

func (r *PasswordResetTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	var model models.PasswordResetTokenModel
	err := r.repository.Executor(ctx).GetContext(ctx, &model, getPasswordResetTokenByHashQuery, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, identityout.ErrPasswordResetTokenNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *PasswordResetTokenRepository) GetLatestActiveForUser(ctx context.Context, userID string, now time.Time) (*domain.PasswordResetToken, error) {
	var model models.PasswordResetTokenModel
	err := r.repository.Executor(ctx).GetContext(ctx, &model, getLatestActivePasswordResetTokenQuery, userID, now)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *PasswordResetTokenRepository) CountCreatedSince(ctx context.Context, userID string, since time.Time) (int, error) {
	var count int
	err := r.repository.Executor(ctx).GetContext(ctx, &count, countPasswordResetTokensQuery, userID, since)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PasswordResetTokenRepository) InvalidateActiveForUser(ctx context.Context, userID string, at time.Time) error {
	_, err := r.repository.Executor(ctx).ExecContext(ctx, invalidatePasswordResetTokensForUserQuery, userID, at)
	return err
}

func (r *PasswordResetTokenRepository) InvalidateByID(ctx context.Context, id string, at time.Time) error {
	_, err := r.repository.Executor(ctx).ExecContext(ctx, invalidatePasswordResetTokenByIDQuery, id, at)
	return err
}

// Consume is the authoritative, atomic redemption check. A single UPDATE
// both validates and marks the token terminal, so two concurrent resets with
// the same token cannot both succeed.
func (r *PasswordResetTokenRepository) Consume(ctx context.Context, tokenHash, userID string, now time.Time) (*domain.PasswordResetToken, error) {
	var model models.PasswordResetTokenModel
	err := r.repository.Executor(ctx).QueryRowxContext(
		ctx,
		consumePasswordResetTokenQuery,
		tokenHash,
		userID,
		now,
	).StructScan(&model)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, identityout.ErrPasswordResetTokenNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *PasswordResetTokenRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.repository.Executor(ctx).ExecContext(ctx, deletePasswordResetTokensQuery, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

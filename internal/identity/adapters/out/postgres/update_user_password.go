package identitypostgres

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

var updateUserPasswordQuery = `
	UPDATE users
	SET password_hash = :password_hash,
		updated_at = NOW()
	WHERE id = :id
`

func (r *UserRepository) UpdateUserPassword(ctx context.Context, user *domain.User) error {
	executor := r.repository.Executor(ctx)

	result, err := executor.NamedExecContext(ctx, updateUserPasswordQuery, map[string]any{
		"id":            user.ID(),
		"password_hash": user.PasswordHash().Value(),
	})
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

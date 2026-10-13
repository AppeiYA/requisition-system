package identitypostgres

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

var updateUserQuery = `
	UPDATE users
	SET first_name = :first_name,
		last_name = :last_name,
		email = :email,
		status = :status,
		updated_at = NOW()
	WHERE id = :id
`

func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	executor := r.repository.Executor(ctx)

	result, err := executor.NamedExecContext(ctx, updateUserQuery, map[string]any{
		"id":         user.ID(),
		"first_name": user.FirstName().Value(),
		"last_name":  user.LastName().Value(),
		"email":      user.Email().Value(),
		"status":     user.Status().Value(),
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

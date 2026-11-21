package identitypostgres

import (
	"context"
	"fmt"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

var updateUserStatusQuery = `
	UPDATE users
	SET status = :status,
		updated_at = NOW()
	WHERE id = :id
`

func (r *UserRepository) UpdateUserStatus(ctx context.Context, user *domain.User) error {
	key := fmt.Sprintf(`%s%s`, UserKeyPrefix, user.ID())
	executor := r.repository.Executor(ctx)

	result, err := executor.NamedExecContext(ctx, updateUserStatusQuery, map[string]any{
		"id":     user.ID(),
		"status": user.Status().Value(),
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

	err = r.cache.Delete(ctx, key)
	if err != nil {
		return err
	}

	return nil
}

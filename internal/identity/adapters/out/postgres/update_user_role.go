package identitypostgres

import (
	"context"
	"fmt"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

var updateUserRoleQuery = `
	UPDATE users
	SET role = :role,
		updated_at = NOW()
	WHERE id = :id
`

func (r *UserRepository) UpdateUserRole(ctx context.Context, user *domain.User) error {
	executor := r.repository.Executor(ctx)

	result, err := executor.NamedExecContext(ctx, updateUserRoleQuery, map[string]any{
		"id":   user.ID(),
		"role": user.Role().String(),
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

	// The auth middleware reads the cached user on every request. Leaving a
	// stale role in the cache would keep a demoted admin authorised.
	return r.cache.Delete(ctx, fmt.Sprintf("%s%s", UserKeyPrefix, user.ID()))
}

package identitypostgres

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

var updateUserQuery = `
	UPDATE users
	SET first_name = :first_name,
		last_name = :last_name,
		phone_number = :phone_number,
		staff_id = :staff_id,
		email = :email,
		status = :status,
		updated_at = NOW()
	WHERE id = :id
`

func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	executor := r.repository.Executor(ctx)

	var phone *string
	if value := user.Phone().Value(); value != "" {
		phone = &value
	}

	var staffID *string
	if value := user.StaffID(); value != "" {
		staffID = &value
	}

	result, err := executor.NamedExecContext(ctx, updateUserQuery, map[string]any{
		"id":           user.ID(),
		"first_name":   user.FirstName().Value(),
		"last_name":    user.LastName().Value(),
		"phone_number": phone,
		"staff_id":     staffID,
		"email":        user.Email().Value(),
		"status":       user.Status().Value(),
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

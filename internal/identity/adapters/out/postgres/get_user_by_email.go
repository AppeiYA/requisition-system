package identitypostgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres/models"
	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var ErrUserNotFound = customerrors.NewNotFoundError("user not found")

var getUserByEmailQuery = `
	SELECT id, first_name, last_name, phone_number, staff_id, email, password_hash, status, role, created_at, updated_at
	FROM users
	WHERE LOWER(email) = LOWER($1)
`

func (r *UserRepository) GetUserByEmail(ctx context.Context, email *shareddomain.Email) (*domain.User, error) {
	var userModel models.UserModel
	var executor = r.repository.Executor(ctx)
	err := executor.GetContext(ctx, &userModel, getUserByEmailQuery, email.Value())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return userModel.ToDomain()
}

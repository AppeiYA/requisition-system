package identitypostgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres/models"
	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

var getUserByIDQuery = `
	SELECT id, first_name, last_name, phone_number, staff_id, email, password_hash, status, created_at, updated_at
	FROM users
	WHERE id = $1
`

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	var userModel models.UserModel
	var executor = r.repository.Executor(ctx)
	err := executor.GetContext(ctx, &userModel, getUserByIDQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return userModel.ToDomain()
}

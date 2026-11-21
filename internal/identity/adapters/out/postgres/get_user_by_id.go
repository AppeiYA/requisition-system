package identitypostgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres/models"
	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

var UserKeyPrefix = "user:"
var UserCacheExpiration = 5 * time.Minute

var getUserByIDQuery = `
	SELECT id, first_name, last_name, phone_number, staff_id, email, password_hash, status, role, created_at, updated_at
	FROM users
	WHERE id = $1
`

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	var userModel models.UserModel
	key := fmt.Sprintf("%s%s", UserKeyPrefix, id)
	found, err := r.cache.Get(ctx, key, &userModel)
	if err != nil {
		return nil, err
	}
	if found {
		return userModel.ToDomain()
	}

	var executor = r.repository.Executor(ctx)
	err = executor.GetContext(ctx, &userModel, getUserByIDQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	err = r.cache.Set(ctx, key, userModel, UserCacheExpiration)
	if err != nil {
		return nil, err
	}
	return userModel.ToDomain()
}

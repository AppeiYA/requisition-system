package identitypostgres

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres/models"
	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

var saveUserQuery = `
	INSERT INTO users (id, first_name, last_name, phone_number, staff_id, email, password_hash, status, created_at, updated_at)
	VALUES (:id, :first_name, :last_name, :phone_number, :staff_id, :email, :password_hash, :status, :created_at, :updated_at)
`

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	_, err := r.repository.Executor(ctx).NamedExecContext(ctx, saveUserQuery, models.FromDomain(user))
	if err != nil {
		return err
	}
	return nil
}

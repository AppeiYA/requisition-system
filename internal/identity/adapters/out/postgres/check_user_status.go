package identitypostgres

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres/models"
	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

func (r *UserRepository) CheckUserStatus(ctx context.Context, id string) (domain.Status, error) {
	var model models.UserStatusModel
	var executor = r.repository.Executor(ctx)
	err := executor.GetContext(ctx, &model, "SELECT id, status, updated_at FROM users WHERE id = $1", id)
	if err != nil {
		return "", err
	}
	return model.ToDomain()
}
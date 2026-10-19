package identitypostgres

import (
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
)

type UserRepository struct {
	repository db.Repository
}

func NewUserRepository(repository db.Repository) *UserRepository {
	return &UserRepository{
		repository: repository,
	}
}

var _ identityout.UsersRepository = (*UserRepository)(nil)

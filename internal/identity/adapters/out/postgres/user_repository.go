package identitypostgres

import (
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/cache"
)

type UserRepository struct {
	repository db.Repository
	cache cache.Cache
}

func NewUserRepository(repository db.Repository, cache cache.Cache) *UserRepository {
	return &UserRepository{
		repository: repository,
		cache: cache,
	}
}

var _ identityout.UsersRepository = (*UserRepository)(nil)

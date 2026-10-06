package identityout

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

type UsersRepository interface {
	SaveUser(ctx context.Context, user *domain.User) error
	GetUserByEmail(ctx context.Context, email *domain.Email) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	UpdateUserStatus(ctx context.Context, user *domain.User) error
	UpdateUserPassword(ctx context.Context, user *domain.User) error
	UpdateUser(ctx context.Context, user *domain.User) error
}
package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

const (
	defaultUserListLimit = 20
	maxUserListLimit     = 100
)

type ListUsersUsecase struct {
	users identityout.UsersRepository
}

func NewListUsersUsecase(users identityout.UsersRepository) *ListUsersUsecase {
	return &ListUsersUsecase{users: users}
}

func (uc *ListUsersUsecase) Execute(ctx context.Context, filter domain.UserListFilter) (*identityin.ListUsersResult, error) {
	if !domain.ValidUserSortValues[filter.Sort] {
		return nil, ErrInvalidUserSort
	}
	if filter.Limit <= 0 {
		filter.Limit = defaultUserListLimit
	}
	if filter.Limit > maxUserListLimit {
		filter.Limit = maxUserListLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	users, err := uc.users.ListUsers(ctx, filter)
	if err != nil {
		return nil, err
	}
	total, err := uc.users.CountUsers(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &identityin.ListUsersResult{Users: users, Total: total}, nil
}

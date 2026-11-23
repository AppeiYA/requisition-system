package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

type GetUserUsecase struct {
	users identityout.UsersRepository
}

func NewGetUserUsecase(users identityout.UsersRepository) *GetUserUsecase {
	return &GetUserUsecase{users: users}
}

func (uc *GetUserUsecase) Execute(ctx context.Context, userID string) (*domain.UserDetail, error) {
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	detail, err := uc.users.GetUserDetail(ctx, userID)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, ErrUserNotFound
	}
	return detail, nil
}

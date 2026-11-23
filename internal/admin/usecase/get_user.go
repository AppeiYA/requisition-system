package usecase

import (
	"context"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
)

type GetUserUsecase struct{ users adminout.UserAdminPort }

func NewGetUserUsecase(users adminout.UserAdminPort) *GetUserUsecase {
	return &GetUserUsecase{users: users}
}

func (uc *GetUserUsecase) Execute(ctx context.Context, userID string) (*adminout.UserDetail, error) {
	return uc.users.GetUser(ctx, userID)
}

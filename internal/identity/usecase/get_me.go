package usecase

import (
	"context"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

type GetMeUsecase struct {
	userRepo identityout.UsersRepository
}

func NewGetMeUsecase(userRepo identityout.UsersRepository, tokenSvc identityout.TokenService) *GetMeUsecase {
	return &GetMeUsecase{
		userRepo: userRepo,
	}
}

func (uc *GetMeUsecase) Execute(ctx context.Context, userID string) (*identityin.UserResponse, error) {
	// get user by ID
	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// create response
	response := &identityin.UserResponse{
		ID:        user.ID(),
		Email:     user.Email().Value(),
		FirstName: user.FirstName().Value(),
		LastName:  user.LastName().Value(),
		CreatedAt: user.CreatedAt(),
		UpdatedAt: user.UpdatedAt(),
	}

	return response, nil
}
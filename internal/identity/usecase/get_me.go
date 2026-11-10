package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

type GetMeUsecase struct {
	userRepo identityout.UsersRepository
}

func NewGetMeUsecase(userRepo identityout.UsersRepository) *GetMeUsecase {
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

	if !user.IsActive() {
		return nil, domain.ErrInactiveUser
	}

	// create response
	response := &identityin.UserResponse{
		ID:        user.ID(),
		Email:     user.Email().Value(),
		FirstName: user.FirstName().Value(),
		LastName:  user.LastName().Value(),
		Phone:     user.Phone().Value(),
		StaffID:   user.StaffID(),
		CreatedAt: user.CreatedAt(),
		UpdatedAt: user.UpdatedAt(),
	}

	return response, nil
}

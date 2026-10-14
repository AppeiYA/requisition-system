package usecase

import (
	"context"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

// ErrUserNotFound is returned when no user exists for the requested email.
// It is a 404 CustomError so callers can distinguish "missing" from failures.
var ErrUserNotFound = customerrors.NewNotFoundError("user not found")

type GetUserByEmailUsecase struct {
	userRepo identityout.UsersRepository
}

func NewGetUserByEmailUsecase(userRepo identityout.UsersRepository) *GetUserByEmailUsecase {
	return &GetUserByEmailUsecase{
		userRepo: userRepo,
	}
}

func (uc *GetUserByEmailUsecase) Execute(ctx context.Context, email string) (*identityin.UserResponse, error) {
	validEmail, err := shareddomain.NewEmail(email)
	if err != nil {
		return nil, err
	}

	user, err := uc.userRepo.GetUserByEmail(ctx, validEmail)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

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

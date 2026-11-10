package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

type CheckUserStatusUsecase struct {
	userRepository identityout.UsersRepository
}

func NewCheckUserStatusUsecase(userRepository identityout.UsersRepository) *CheckUserStatusUsecase {
	return &CheckUserStatusUsecase{
		userRepository: userRepository,
	}
}

func (uc *CheckUserStatusUsecase) Execute(ctx context.Context ,userID string) (domain.Status, error) {
	userStatus, err := uc.userRepository.CheckUserStatus(ctx, userID)
	if err != nil {
		return "", err
	}
	return userStatus, nil
}
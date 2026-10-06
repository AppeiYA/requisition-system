package usecase

import identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"

type RegisterUserUsecase struct {
	userRepo identityout.UsersRepository
}
package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

type LoginUserUsecase struct {
	userRepository identityout.UsersRepository
	tokenService   identityout.TokenService
	passwordHasher identityout.PasswordHasher
}

func NewLoginUserUsecase(userRepository identityout.UsersRepository, tokenService identityout.TokenService, passwordHasher identityout.PasswordHasher) *LoginUserUsecase {
	return &LoginUserUsecase{
		userRepository: userRepository,
		tokenService:   tokenService,
		passwordHasher: passwordHasher,
	}
}

func (u *LoginUserUsecase) Execute(ctx context.Context, req identityin.LoginUserRequest) (string, error) {
	// create new email vo
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return "", err
	}

	// get user by email
	user, err := u.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	// create new password vo
	password, err := domain.NewPassword(req.Password)
	if err != nil {
		return "", err
	}

	// compare password
	err = u.passwordHasher.ComparePassword(user.PasswordHash().Value(), password.Value())
	if err != nil {
		return "", err
	}

	// generate token
	token, err := u.tokenService.GenerateToken(ctx, &identityout.TokenModel{
		ID:    user.ID(),
		Email: user.Email().Value(),
	})
	
	if err != nil {
		return "", err
	}

	return token, nil
}
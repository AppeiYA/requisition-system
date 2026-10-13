package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

type LoginUserUsecase struct {
	userRepository        identityout.UsersRepository
	accessTokenService    identityout.AccessTokenService
	refreshSessionRepo    identityout.RefreshSessionRepository
	refreshTokenGenerator identityout.RefreshTokenGenerator
	passwordHasher        identityout.PasswordHasher
}

func NewLoginUserUsecase(
	userRepository identityout.UsersRepository,
	accessTokenService identityout.AccessTokenService,
	refreshSessionRepo identityout.RefreshSessionRepository,
	refreshTokenGenerator identityout.RefreshTokenGenerator,
	passwordHasher identityout.PasswordHasher,
) *LoginUserUsecase {
	return &LoginUserUsecase{
		userRepository:        userRepository,
		accessTokenService:    accessTokenService,
		refreshSessionRepo:    refreshSessionRepo,
		refreshTokenGenerator: refreshTokenGenerator,
		passwordHasher:        passwordHasher,
	}
}

func (u *LoginUserUsecase) Execute(ctx context.Context, req identityin.LoginUserRequest) (*identityin.TokenResponse, error) {
	// create new email vo
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return nil, err
	}

	// get user by email
	user, err := u.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// create new password vo
	password, err := domain.NewPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// compare password
	if err := u.passwordHasher.ComparePassword(user.PasswordHash().Value(), password.Value()); err != nil {
		return nil, err
	}

	// issue a short-lived access token
	access, err := u.accessTokenService.GenerateAccessToken(ctx, &identityout.TokenModel{ID: user.ID()})
	if err != nil {
		return nil, err
	}

	// issue an opaque refresh token and persist only its hash
	refreshToken, err := u.refreshTokenGenerator.Generate()
	if err != nil {
		return nil, err
	}

	if err := u.refreshSessionRepo.Create(ctx, user.ID(), u.refreshTokenGenerator.Hash(refreshToken)); err != nil {
		return nil, err
	}

	return newTokenResponse(access, refreshToken), nil
}

package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var ErrInvalidCredentials = customerrors.NewUnauthorizedError("invalid email or password")

type LoginUserUsecase struct {
	userRepository        identityout.UsersRepository
	accessTokenService    identityout.AccessTokenService
	refreshSessionRepo    identityout.RefreshSessionRepository
	refreshTokenGenerator identityout.RefreshTokenGenerator
	passwordHasher        identityout.PasswordHasher
	dummyPasswordHash     string
}

func NewLoginUserUsecase(
	userRepository identityout.UsersRepository,
	accessTokenService identityout.AccessTokenService,
	refreshSessionRepo identityout.RefreshSessionRepository,
	refreshTokenGenerator identityout.RefreshTokenGenerator,
	passwordHasher identityout.PasswordHasher,
) *LoginUserUsecase {
	dummyPasswordHash, _ := passwordHasher.HashPassword("DummyPassword123$")
	return &LoginUserUsecase{
		userRepository:        userRepository,
		accessTokenService:    accessTokenService,
		refreshSessionRepo:    refreshSessionRepo,
		refreshTokenGenerator: refreshTokenGenerator,
		passwordHasher:        passwordHasher,
		dummyPasswordHash:     dummyPasswordHash,
	}
}

func (u *LoginUserUsecase) Execute(ctx context.Context, req identityin.LoginUserRequest) (*identityin.TokenResponse, error) {
	// create new email vo
	email, err := shareddomain.NewEmail(req.Email)
	if err != nil {
		return nil, err
	}

	password, err := domain.NewPassword(req.Password)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// get user by email
	user, err := u.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		if err, ok := err.(*customerrors.CustomError); ok && err.Code == 404 {
			_ = u.passwordHasher.ComparePassword(u.dummyPasswordHash, password.Value())
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.IsActive() {
		_ = u.passwordHasher.ComparePassword(user.PasswordHash().Value(), password.Value())
		return nil, ErrInvalidCredentials
	}

	// compare password
	if err := u.passwordHasher.ComparePassword(user.PasswordHash().Value(), password.Value()); err != nil {
		return nil, ErrInvalidCredentials
	}

	// issue a short-lived access token
	access, err := u.accessTokenService.GenerateAccessToken(ctx, &identityout.TokenModel{
		ID:    user.ID(),
		Role:  user.Role().String(),
		Email: user.Email().Value(),
	})
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

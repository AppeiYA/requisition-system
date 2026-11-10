package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

type RefreshTokenUsecase struct {
	accessTokenService    identityout.AccessTokenService
	refreshSessionRepo    identityout.RefreshSessionRepository
	refreshTokenGenerator identityout.RefreshTokenGenerator
	userRepo 			  identityout.UsersRepository
}

func NewRefreshTokenUsecase(
	accessTokenService identityout.AccessTokenService,
	refreshSessionRepo identityout.RefreshSessionRepository,
	refreshTokenGenerator identityout.RefreshTokenGenerator,
	userRepo identityout.UsersRepository,
) *RefreshTokenUsecase {
	return &RefreshTokenUsecase{
		accessTokenService:    accessTokenService,
		refreshSessionRepo:    refreshSessionRepo,
		refreshTokenGenerator: refreshTokenGenerator,
		userRepo: userRepo,
	}
}

// Execute rotates a refresh token: the presented token is atomically consumed
// and replaced, and a new access token is issued. The session repository
// guarantees that a token can be consumed only once.
func (u *RefreshTokenUsecase) Execute(
	ctx context.Context,
	req identityin.RefreshTokenRequest,
) (*identityin.TokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, customerrors.NewBadRequestError("refresh token is required")
	}

	newRefreshToken, err := u.refreshTokenGenerator.Generate()
	if err != nil {
		return nil, err
	}

	session, err := u.refreshSessionRepo.Rotate(
		ctx,
		u.refreshTokenGenerator.Hash(req.RefreshToken),
		u.refreshTokenGenerator.Hash(newRefreshToken),
	)
	if err != nil {
		return nil, err
	}

	user, err := u.userRepo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive() {
		return nil, domain.ErrInactiveUser
	}

	access, err := u.accessTokenService.GenerateAccessToken(
		ctx, 
		&identityout.TokenModel{
			ID: session.UserID,
			Role: user.Role().String(),
			Email: user.Email().Value(),
		},
	)
	if err != nil {
		return nil, err
	}

	return newTokenResponse(access, newRefreshToken), nil
}

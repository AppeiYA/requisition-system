package usecase

import (
	"context"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/errors"
)

type RefreshTokenUsecase struct {
	accessTokenService    identityout.AccessTokenService
	refreshSessionRepo    identityout.RefreshSessionRepository
	refreshTokenGenerator identityout.RefreshTokenGenerator
}

func NewRefreshTokenUsecase(
	accessTokenService identityout.AccessTokenService,
	refreshSessionRepo identityout.RefreshSessionRepository,
	refreshTokenGenerator identityout.RefreshTokenGenerator,
) *RefreshTokenUsecase {
	return &RefreshTokenUsecase{
		accessTokenService:    accessTokenService,
		refreshSessionRepo:    refreshSessionRepo,
		refreshTokenGenerator: refreshTokenGenerator,
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

	access, err := u.accessTokenService.GenerateAccessToken(ctx, &identityout.TokenModel{ID: session.UserID})
	if err != nil {
		return nil, err
	}

	return newTokenResponse(access, newRefreshToken), nil
}

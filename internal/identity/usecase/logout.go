package usecase

import (
	"context"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/errors"
)

type LogoutUsecase struct {
	refreshSessionRepo    identityout.RefreshSessionRepository
	refreshTokenGenerator identityout.RefreshTokenGenerator
}

func NewLogoutUsecase(
	refreshSessionRepo identityout.RefreshSessionRepository,
	refreshTokenGenerator identityout.RefreshTokenGenerator,
) *LogoutUsecase {
	return &LogoutUsecase{
		refreshSessionRepo:    refreshSessionRepo,
		refreshTokenGenerator: refreshTokenGenerator,
	}
}

// Execute revokes the refresh session for the presented token and its whole
// family. Access tokens are short-lived and are not tracked, so they simply
// expire.
func (u *LogoutUsecase) Execute(ctx context.Context, req identityin.LogoutRequest) error {
	if req.RefreshToken == "" {
		return customerrors.NewBadRequestError("refresh token is required")
	}

	return u.refreshSessionRepo.Revoke(ctx, u.refreshTokenGenerator.Hash(req.RefreshToken))
}

package usecase

import (
	"context"
	"time"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
)

// CleanupPasswordResetTokensUsecase removes rows created before the retention
// window. It runs on a schedule in the worker process.
type CleanupPasswordResetTokensUsecase struct {
	tokenRepo identityout.PasswordResetTokenRepository
	retention time.Duration
}

func NewCleanupPasswordResetTokensUsecase(
	tokenRepo identityout.PasswordResetTokenRepository,
	cfg config.PasswordResetConfig,
) *CleanupPasswordResetTokensUsecase {
	retention := cfg.CleanupRetention
	if retention <= 0 {
		retention = 7 * 24 * time.Hour
	}

	return &CleanupPasswordResetTokensUsecase{
		tokenRepo: tokenRepo,
		retention: retention,
	}
}

func (u *CleanupPasswordResetTokensUsecase) Execute(ctx context.Context) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	return u.tokenRepo.DeleteOlderThan(ctx, time.Now().UTC().Add(-u.retention))
}

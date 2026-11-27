package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

// ResetPasswordUsecase redeems a reset token and replaces the user's password.
// It never issues credentials: the user signs in normally afterwards.
type ResetPasswordUsecase struct {
	userRepo           identityout.UsersRepository
	tokenRepo          identityout.PasswordResetTokenRepository
	refreshSessionRepo identityout.RefreshSessionRepository
	tokenGenerator     identityout.OpaqueTokenGenerator
	passwordHasher     identityout.PasswordHasher
	txManager          identityout.TransactionManager
	telemetry          identityout.Telemetry
}

func NewResetPasswordUsecase(
	userRepo identityout.UsersRepository,
	tokenRepo identityout.PasswordResetTokenRepository,
	refreshSessionRepo identityout.RefreshSessionRepository,
	tokenGenerator identityout.OpaqueTokenGenerator,
	passwordHasher identityout.PasswordHasher,
	txManager identityout.TransactionManager,
	telemetry identityout.Telemetry,
) *ResetPasswordUsecase {
	return &ResetPasswordUsecase{
		userRepo:           userRepo,
		tokenRepo:          tokenRepo,
		refreshSessionRepo: refreshSessionRepo,
		tokenGenerator:     tokenGenerator,
		passwordHasher:     passwordHasher,
		txManager:          txManager,
		telemetry:          telemetry,
	}
}

func (u *ResetPasswordUsecase) Execute(ctx context.Context, token, password string) error {
	return u.telemetry.Measure(ctx, identityout.OperationResetPassword, func(ctx context.Context) error {
		return u.execute(ctx, token, password)
	})
}

func (u *ResetPasswordUsecase) execute(ctx context.Context, token, password string) error {
	if token == "" {
		u.logRejected(ctx, "invalid_token")
		return ErrPasswordResetTokenRequired
	}

	// Enforce the existing policy before touching storage.
	validPassword, err := domain.NewPassword(password)
	if err != nil {
		u.logRejected(ctx, "weak_password")
		return err
	}
	hashedPassword, err := u.passwordHasher.HashPassword(validPassword.Value())
	if err != nil {
		return err
	}
	newHash, err := domain.NewPasswordHash(hashedPassword)
	if err != nil {
		return err
	}

	tokenHash := u.tokenGenerator.Hash(token)
	now := time.Now().UTC()

	// Read-only pre-check: it only exists to learn the user ID before taking
	// the user row lock. Consume is the authoritative check.
	savedToken, err := u.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, identityout.ErrPasswordResetTokenNotFound) || isNotFoundError(err) {
			u.logRejected(ctx, "invalid_token")
			return ErrInvalidPasswordResetToken
		}
		return err
	}
	if savedToken == nil || !savedToken.IsActive(now) {
		u.logRejected(ctx, "invalid_token")
		return ErrInvalidPasswordResetToken
	}

	reason := "invalid_token"
	err = u.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Lock order is always user row first, then token row.
		user, err := u.userRepo.GetUserByIDForUpdate(txCtx, savedToken.UserID())
		if err != nil {
			if isNotFoundError(err) {
				return ErrInvalidPasswordResetToken
			}
			return err
		}
		if user == nil {
			return ErrInvalidPasswordResetToken
		}
		if !user.IsActive() {
			reason = "inactive_user"
			return ErrInvalidPasswordResetToken
		}

		consumed, err := u.tokenRepo.Consume(txCtx, tokenHash, user.ID(), time.Now().UTC())
		if err != nil {
			if errors.Is(err, identityout.ErrPasswordResetTokenNotFound) || isNotFoundError(err) {
				return ErrInvalidPasswordResetToken
			}
			return err
		}
		if consumed == nil {
			return ErrInvalidPasswordResetToken
		}

		if err := user.ChangePassword(newHash); err != nil {
			return err
		}
		if err := u.userRepo.UpdateUserPassword(txCtx, user); err != nil {
			return err
		}
		// Defence in depth: Consume made this token terminal, but no other
		// redeemable token may survive a completed reset.
		if err := u.tokenRepo.InvalidateActiveForUser(txCtx, user.ID(), time.Now().UTC()); err != nil {
			return err
		}
		// Fail closed: if Redis is unavailable the transaction rolls back and
		// the token remains redeemable. Old sessions must not outlive the
		// password change.
		if err := u.refreshSessionRepo.RevokeAll(txCtx, user.ID()); err != nil {
			return err
		}

		reason = ""
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrInvalidPasswordResetToken) || errors.Is(err, ErrPasswordResetTokenRequired) {
			u.logRejected(ctx, reason)
			return ErrInvalidPasswordResetToken
		}
		return err
	}

	slog.InfoContext(ctx, "password_reset.completed",
		slog.String("event", "password_reset.completed"),
		slog.String("user.id", savedToken.UserID()),
	)
	return nil
}

func (u *ResetPasswordUsecase) logRejected(ctx context.Context, reason string) {
	if reason == "" {
		reason = "invalid_token"
	}
	slog.InfoContext(ctx, "password_reset.rejected",
		slog.String("event", "password_reset.rejected"),
		slog.String("reason", reason),
	)
}

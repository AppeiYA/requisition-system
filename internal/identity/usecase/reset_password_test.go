package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identitymocks "github.com/AppeiYA/requisition-system/internal/identity/mocks"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

func newResetPasswordUsecase(
	users identityout.UsersRepository,
	tokens identityout.PasswordResetTokenRepository,
	sessions identityout.RefreshSessionRepository,
	tx identityout.TransactionManager,
	telemetry identityout.Telemetry,
) *ResetPasswordUsecase {
	return NewResetPasswordUsecase(
		users,
		tokens,
		sessions,
		identitymocks.NewMockRefreshTokenGenerator(),
		identitymocks.NewMockPasswordHasher(),
		tx,
		telemetry,
	)
}

func validResetToken(t *testing.T, userID string) *domain.PasswordResetToken {
	t.Helper()

	token, err := domain.NewPasswordResetToken(
		"prt-1",
		userID,
		"hash:reset-token",
		time.Now().UTC(),
		30*time.Minute,
	)
	if err != nil {
		t.Fatalf("create reset token: %v", err)
	}
	return token
}

func Test_ResetPasswordUsecase_Success(t *testing.T) {
	t.Parallel()

	token := validResetToken(t, identitymocks.DefaultID)
	tokens := identitymocks.NewMockPasswordResetTokenRepository()
	tokens.GetByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
		if tokenHash != token.TokenHash() {
			t.Errorf("lookup hash = %q, want %q", tokenHash, token.TokenHash())
		}
		return token, nil
	}
	consumeCalled := false
	tokens.ConsumeFunc = func(ctx context.Context, tokenHash, userID string, now time.Time) (*domain.PasswordResetToken, error) {
		consumeCalled = true
		if tokenHash != token.TokenHash() || userID != identitymocks.DefaultID {
			t.Errorf("consume args = (%q, %q)", tokenHash, userID)
		}
		return token, nil
	}

	users := identitymocks.NewMockUsersRepository()
	var updated *domain.User
	users.UpdateUserPasswordFunc = func(ctx context.Context, user *domain.User) error {
		updated = user
		return nil
	}
	users.GetUserByIDForUpdateFunc = func(ctx context.Context, id string) (*domain.User, error) {
		return identitymocks.ValidUser(), nil
	}

	sessions := identitymocks.NewMockRefreshSessionRepository()
	var revokedUserID string
	sessions.RevokeAllFunc = func(ctx context.Context, userID string) error {
		revokedUserID = userID
		return nil
	}

	uc := newResetPasswordUsecase(users, tokens, sessions, identitymocks.NewMockTransactionManager(), identitymocks.NewMockTelemetry())
	if err := uc.Execute(context.Background(), "reset-token", "NewPassword123$"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !consumeCalled {
		t.Fatalf("token was not consumed")
	}
	if updated == nil || updated.PasswordHash().Value() != "NewPassword123$" {
		t.Fatalf("password hash was not updated: %+v", updated)
	}
	if updated.UpdatedAt() == nil {
		t.Fatalf("updatedAt was not set")
	}
	if revokedUserID != identitymocks.DefaultID {
		t.Fatalf("revoked user = %q, want %q", revokedUserID, identitymocks.DefaultID)
	}
}

func Test_ResetPasswordUsecase_RejectsWeakPasswordBeforeStorage(t *testing.T) {
	t.Parallel()

	tokens := identitymocks.NewMockPasswordResetTokenRepository()
	tokens.GetByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
		t.Errorf("token storage must not be touched for a weak password")
		return nil, identityout.ErrPasswordResetTokenNotFound
	}
	sessions := identitymocks.NewMockRefreshSessionRepository()
	sessions.RevokeAllFunc = func(ctx context.Context, userID string) error {
		t.Errorf("sessions must not be touched for a weak password")
		return nil
	}

	uc := newResetPasswordUsecase(
		identitymocks.NewMockUsersRepository(),
		tokens,
		sessions,
		identitymocks.NewMockTransactionManager(),
		identitymocks.NewMockTelemetry(),
	)
	err := uc.Execute(context.Background(), "reset-token", "weak")
	if err == nil {
		t.Fatalf("expected weak password error")
	}
}

func Test_ResetPasswordUsecase_GenericErrorForBadTokensAndInactiveUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("unknown token", func(t *testing.T) {
		tokens := identitymocks.NewMockPasswordResetTokenRepository()
		uc := newResetPasswordUsecase(identitymocks.NewMockUsersRepository(), tokens, identitymocks.NewMockRefreshSessionRepository(), identitymocks.NewMockTransactionManager(), identitymocks.NewMockTelemetry())
		err := uc.Execute(ctx, "reset-token", "NewPassword123$")
		if !errors.Is(err, ErrInvalidPasswordResetToken) {
			t.Fatalf("error = %v, want %v", err, ErrInvalidPasswordResetToken)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		token := validResetToken(t, identitymocks.DefaultID)
		expired, err := domain.NewPasswordResetToken(token.ID(), token.UserID(), token.TokenHash(), time.Now().UTC().Add(-time.Hour), time.Minute)
		if err != nil {
			t.Fatalf("expired token: %v", err)
		}
		tokens := identitymocks.NewMockPasswordResetTokenRepository()
		tokens.GetByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
			return expired, nil
		}
		uc := newResetPasswordUsecase(identitymocks.NewMockUsersRepository(), tokens, identitymocks.NewMockRefreshSessionRepository(), identitymocks.NewMockTransactionManager(), identitymocks.NewMockTelemetry())
		if err := uc.Execute(ctx, "reset-token", "NewPassword123$"); !errors.Is(err, ErrInvalidPasswordResetToken) {
			t.Fatalf("error = %v, want %v", err, ErrInvalidPasswordResetToken)
		}
	})

	t.Run("inactive user", func(t *testing.T) {
		token := validResetToken(t, identitymocks.DefaultID)
		tokens := identitymocks.NewMockPasswordResetTokenRepository()
		tokens.GetByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
			return token, nil
		}
		user := identitymocks.ValidUser()
		if err := user.UpdateStatus(domain.SUSPENDED); err != nil {
			t.Fatalf("suspend user: %v", err)
		}
		users := identitymocks.NewMockUsersRepository()
		users.GetUserByIDForUpdateFunc = func(ctx context.Context, id string) (*domain.User, error) {
			return user, nil
		}
		sessions := identitymocks.NewMockRefreshSessionRepository()
		sessions.RevokeAllFunc = func(ctx context.Context, userID string) error {
			t.Errorf("sessions must not be revoked for inactive user")
			return nil
		}
		uc := newResetPasswordUsecase(users, tokens, sessions, identitymocks.NewMockTransactionManager(), identitymocks.NewMockTelemetry())
		if err := uc.Execute(ctx, "reset-token", "NewPassword123$"); !errors.Is(err, ErrInvalidPasswordResetToken) {
			t.Fatalf("error = %v, want %v", err, ErrInvalidPasswordResetToken)
		}
	})
}

func Test_ResetPasswordUsecase_RevokeAllFailureFailsTheReset(t *testing.T) {
	t.Parallel()

	token := validResetToken(t, identitymocks.DefaultID)
	tokens := identitymocks.NewMockPasswordResetTokenRepository()
	tokens.GetByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
		return token, nil
	}
	tokens.ConsumeFunc = func(ctx context.Context, tokenHash, userID string, now time.Time) (*domain.PasswordResetToken, error) {
		return token, nil
	}
	users := identitymocks.NewMockUsersRepository()
	users.GetUserByIDForUpdateFunc = func(ctx context.Context, id string) (*domain.User, error) {
		return identitymocks.ValidUser(), nil
	}
	sessions := identitymocks.NewMockRefreshSessionRepository()
	sessions.RevokeAllFunc = func(ctx context.Context, userID string) error {
		return errors.New("redis unavailable")
	}

	uc := newResetPasswordUsecase(users, tokens, sessions, identitymocks.NewMockTransactionManager(), identitymocks.NewMockTelemetry())
	if err := uc.Execute(context.Background(), "reset-token", "NewPassword123$"); err == nil {
		t.Fatalf("expected an error when session revocation fails")
	}
}

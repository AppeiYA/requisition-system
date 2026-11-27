package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identitymocks "github.com/AppeiYA/requisition-system/internal/identity/mocks"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	identitytasks "github.com/AppeiYA/requisition-system/internal/identity/tasks"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

func newRequestPasswordResetUsecase(
	users identityout.UsersRepository,
	tokens identityout.PasswordResetTokenRepository,
	queue identityout.PasswordResetEmailQueue,
	tx identityout.TransactionManager,
	telemetry identityout.Telemetry,
	cfg config.PasswordResetConfig,
) *RequestPasswordResetUsecase {
	return NewRequestPasswordResetUsecase(
		users,
		tokens,
		queue,
		identitymocks.NewMockRefreshTokenGenerator(),
		identitymocks.NewMockIDGenerator(),
		tx,
		cfg,
		telemetry,
	)
}

func testPasswordResetConfig() config.PasswordResetConfig {
	return config.PasswordResetConfig{
		TokenTTL:         30 * time.Minute,
		RequestCooldown:  time.Minute,
		MaxPerHour:       5,
		CleanupRetention: 7 * 24 * time.Hour,
	}
}

func Test_RequestPasswordResetUsecase_ActiveUserCreatesOneTokenAndEmail(t *testing.T) {
	t.Parallel()

	users := identitymocks.NewMockUsersRepository()
	users.GetUserByEmailFunc = func(ctx context.Context, email *shareddomain.Email) (*domain.User, error) {
		return identitymocks.ValidUser(), nil
	}
	users.GetUserByIDForUpdateFunc = func(ctx context.Context, id string) (*domain.User, error) {
		return identitymocks.ValidUser(), nil
	}

	var created *domain.PasswordResetToken
	tokens := identitymocks.NewMockPasswordResetTokenRepository()
	tokens.CreateFunc = func(ctx context.Context, token *domain.PasswordResetToken) error {
		created = token
		return nil
	}
	tokens.InvalidateActiveForUserFunc = func(ctx context.Context, userID string, at time.Time) error {
		if userID != identitymocks.DefaultID {
			t.Errorf("invalidate user = %q, want %q", userID, identitymocks.DefaultID)
		}
		return nil
	}

	queue := identitymocks.NewMockPasswordResetEmailQueue()
	uc := newRequestPasswordResetUsecase(
		users, tokens, queue,
		identitymocks.NewMockTransactionManager(),
		identitymocks.NewMockTelemetry(),
		testPasswordResetConfig(),
	)

	if err := uc.Execute(context.Background(), identitymocks.DefaultEmail); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if created == nil {
		t.Fatalf("no token was created")
	}
	if created.UserID() != identitymocks.DefaultID {
		t.Fatalf("token user = %q", created.UserID())
	}
	if created.TokenHash() != "hash:refresh-token" {
		t.Fatalf("stored hash = %q, want the generator hash", created.TokenHash())
	}
	if created.IsUsed() {
		t.Fatalf("new token must be redeemable")
	}
	if len(queue.Payloads) != 1 {
		t.Fatalf("payloads = %d, want 1", len(queue.Payloads))
	}
	payload := queue.Payloads[0]
	if payload.ResetID != created.ID() || payload.To != identitymocks.DefaultEmail {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload.Token != "refresh-token" {
		t.Fatalf("raw token = %q, want the generated token", payload.Token)
	}
	if _, err := time.Parse(time.RFC3339, payload.ExpiresAt); err != nil {
		t.Fatalf("payload expiry %q is not RFC3339: %v", payload.ExpiresAt, err)
	}
}

func Test_RequestPasswordResetUsecase_UnknownAndInactiveSendNothing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := testPasswordResetConfig()

	t.Run("unknown email", func(t *testing.T) {
		users := identitymocks.NewMockUsersRepository()
		users.GetUserByEmailFunc = func(ctx context.Context, email *shareddomain.Email) (*domain.User, error) {
			return nil, customerrors.NewNotFoundError("user not found")
		}
		tokens := identitymocks.NewMockPasswordResetTokenRepository()
		queue := identitymocks.NewMockPasswordResetEmailQueue()
		uc := newRequestPasswordResetUsecase(users, tokens, queue, identitymocks.NewMockTransactionManager(), identitymocks.NewMockTelemetry(), cfg)

		if err := uc.Execute(ctx, identitymocks.DefaultEmail); err != nil {
			t.Fatalf("unknown email error = %v, want nil", err)
		}
		if len(queue.Payloads) != 0 {
			t.Fatalf("queued %d emails for unknown account", len(queue.Payloads))
		}
	})

	t.Run("inactive user", func(t *testing.T) {
		user := identitymocks.ValidUser()
		if err := user.UpdateStatus(domain.SUSPENDED); err != nil {
			t.Fatalf("suspend user: %v", err)
		}
		users := identitymocks.NewMockUsersRepository()
		users.GetUserByEmailFunc = func(ctx context.Context, email *shareddomain.Email) (*domain.User, error) {
			return user, nil
		}
		tokens := identitymocks.NewMockPasswordResetTokenRepository()
		queue := identitymocks.NewMockPasswordResetEmailQueue()
		uc := newRequestPasswordResetUsecase(users, tokens, queue, identitymocks.NewMockTransactionManager(), identitymocks.NewMockTelemetry(), cfg)

		if err := uc.Execute(ctx, identitymocks.DefaultEmail); err != nil {
			t.Fatalf("inactive user error = %v, want nil", err)
		}
		if len(queue.Payloads) != 0 {
			t.Fatalf("queued %d emails for inactive account", len(queue.Payloads))
		}
	})
}

func Test_RequestPasswordResetUsecase_CooldownAndHourlyCap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := testPasswordResetConfig()

	t.Run("inside cooldown", func(t *testing.T) {
		users := identitymocks.NewMockUsersRepository()
		users.GetUserByEmailFunc = func(ctx context.Context, email *shareddomain.Email) (*domain.User, error) {
			return identitymocks.ValidUser(), nil
		}
		users.GetUserByIDForUpdateFunc = func(ctx context.Context, id string) (*domain.User, error) {
			return identitymocks.ValidUser(), nil
		}
		tokens := identitymocks.NewMockPasswordResetTokenRepository()
		tokens.GetLatestActiveForUserFunc = func(ctx context.Context, userID string, now time.Time) (*domain.PasswordResetToken, error) {
			return domain.NewPasswordResetToken("prt-old", userID, "old-hash", now.Add(-10*time.Second), 30*time.Minute)
		}
		created := false
		tokens.CreateFunc = func(ctx context.Context, token *domain.PasswordResetToken) error {
			created = true
			return nil
		}
		queue := identitymocks.NewMockPasswordResetEmailQueue()
		uc := newRequestPasswordResetUsecase(users, tokens, queue, identitymocks.NewMockTransactionManager(), identitymocks.NewMockTelemetry(), cfg)

		if err := uc.Execute(ctx, identitymocks.DefaultEmail); err != nil {
			t.Fatalf("throttled request error = %v, want nil", err)
		}
		if created || len(queue.Payloads) != 0 {
			t.Fatalf("throttled request created token = %v, emails = %d", created, len(queue.Payloads))
		}
	})

	t.Run("hourly cap", func(t *testing.T) {
		users := identitymocks.NewMockUsersRepository()
		users.GetUserByEmailFunc = func(ctx context.Context, email *shareddomain.Email) (*domain.User, error) {
			return identitymocks.ValidUser(), nil
		}
		users.GetUserByIDForUpdateFunc = func(ctx context.Context, id string) (*domain.User, error) {
			return identitymocks.ValidUser(), nil
		}
		tokens := identitymocks.NewMockPasswordResetTokenRepository()
		tokens.CountCreatedSinceFunc = func(ctx context.Context, userID string, since time.Time) (int, error) {
			return cfg.MaxPerHour, nil
		}
		created := false
		tokens.CreateFunc = func(ctx context.Context, token *domain.PasswordResetToken) error {
			created = true
			return nil
		}
		queue := identitymocks.NewMockPasswordResetEmailQueue()
		uc := newRequestPasswordResetUsecase(users, tokens, queue, identitymocks.NewMockTransactionManager(), identitymocks.NewMockTelemetry(), cfg)

		if err := uc.Execute(ctx, identitymocks.DefaultEmail); err != nil {
			t.Fatalf("capped request error = %v, want nil", err)
		}
		if created || len(queue.Payloads) != 0 {
			t.Fatalf("capped request created token = %v, emails = %d", created, len(queue.Payloads))
		}
	})
}

func Test_RequestPasswordResetUsecase_EnqueueFailureInvalidatesByID(t *testing.T) {
	t.Parallel()

	users := identitymocks.NewMockUsersRepository()
	users.GetUserByEmailFunc = func(ctx context.Context, email *shareddomain.Email) (*domain.User, error) {
		return identitymocks.ValidUser(), nil
	}
	users.GetUserByIDForUpdateFunc = func(ctx context.Context, id string) (*domain.User, error) {
		return identitymocks.ValidUser(), nil
	}

	tokens := identitymocks.NewMockPasswordResetTokenRepository()
	var invalidatedID string
	tokens.InvalidateByIDFunc = func(ctx context.Context, id string, at time.Time) error {
		invalidatedID = id
		return nil
	}

	queue := identitymocks.NewMockPasswordResetEmailQueue()
	queue.EnqueueFunc = func(ctx context.Context, payload identitytasks.PasswordResetEmailPayload) error {
		return errors.New("enqueue unavailable")
	}

	uc := newRequestPasswordResetUsecase(users, tokens, queue, identitymocks.NewMockTransactionManager(), identitymocks.NewMockTelemetry(), testPasswordResetConfig())
	if err := uc.Execute(context.Background(), identitymocks.DefaultEmail); err != nil {
		t.Fatalf("enqueue failure error = %v, want nil", err)
	}
	if invalidatedID == "" {
		t.Fatalf("token was not invalidated after failed enqueue")
	}
}

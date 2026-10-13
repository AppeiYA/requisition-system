package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identitymocks "github.com/AppeiYA/requisition-system/internal/identity/mocks"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/errors"
)

func newLoginUsecase(
	repo identityout.UsersRepository,
	accessTokens identityout.AccessTokenService,
	sessions identityout.RefreshSessionRepository,
	generator identityout.RefreshTokenGenerator,
) *LoginUserUsecase {
	return NewLoginUserUsecase(
		repo,
		accessTokens,
		sessions,
		generator,
		identitymocks.NewMockPasswordHasher(),
	)
}

func Test_LoginUserUsecase_Success(t *testing.T) {
	t.Parallel()

	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *domain.Email) (*domain.User, error) {
		return identitymocks.ValidUser(), nil
	}

	accessTokens := identitymocks.NewMockAccessTokenService()
	accessTokens.GenerateAccessTokenFunc = func(ctx context.Context, user *identityout.TokenModel) (*identityout.AccessToken, error) {
		if user.ID != identitymocks.DefaultID {
			t.Errorf("expected token model ID %q, got %q", identitymocks.DefaultID, user.ID)
		}
		return &identityout.AccessToken{Value: "test-access-token", ExpiresIn: 900}, nil
	}

	generator := identitymocks.NewMockRefreshTokenGenerator()
	generator.GenerateFunc = func() (string, error) {
		return "test-refresh-token", nil
	}

	var createdUserID, createdTokenHash string
	sessions := identitymocks.NewMockRefreshSessionRepository()
	sessions.CreateFunc = func(ctx context.Context, userID, tokenHash string) error {
		createdUserID = userID
		createdTokenHash = tokenHash
		return nil
	}

	uc := newLoginUsecase(repo, accessTokens, sessions, generator)

	resp, err := uc.Execute(context.Background(), identityin.LoginUserRequest{
		Email:    identitymocks.DefaultEmail,
		Password: identitymocks.DefaultPassword,
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if resp == nil {
		t.Fatal("expected token response, got nil")
	}
	if resp.AccessToken != "test-access-token" {
		t.Errorf("expected access token 'test-access-token', got: %q", resp.AccessToken)
	}
	if resp.RefreshToken != "test-refresh-token" {
		t.Errorf("expected refresh token 'test-refresh-token', got: %q", resp.RefreshToken)
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("expected token type Bearer, got: %q", resp.TokenType)
	}
	if resp.ExpiresIn != 900 {
		t.Errorf("expected ExpiresIn 900, got: %d", resp.ExpiresIn)
	}

	// only the hash of the refresh token must be persisted
	if createdUserID != identitymocks.DefaultID {
		t.Errorf("expected session for user %q, got %q", identitymocks.DefaultID, createdUserID)
	}
	if createdTokenHash != "hash:test-refresh-token" {
		t.Errorf("expected persisted token hash %q, got %q", "hash:test-refresh-token", createdTokenHash)
	}
	if createdTokenHash == resp.RefreshToken {
		t.Error("raw refresh token must not be persisted")
	}
}

func Test_LoginUserUsecase_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		request     identityin.LoginUserRequest
		expectedErr error
	}{
		{
			name: "invalid email",
			request: identityin.LoginUserRequest{
				Email:    "invalid-email",
				Password: identitymocks.DefaultPassword,
			},
			expectedErr: domain.ErrInvalidEmailFormat,
		},
		{
			name: "user not found",
			request: identityin.LoginUserRequest{
				Email:    "unknown@example.com",
				Password: identitymocks.DefaultPassword,
			},
			expectedErr: customerrors.NewNotFoundError("user not found"),
		},
		{
			name: "invalid password format",
			request: identityin.LoginUserRequest{
				Email:    identitymocks.DefaultEmail,
				Password: "short",
			},
			expectedErr: domain.ErrInvalidPassword,
		},
		{
			name: "wrong password",
			request: identityin.LoginUserRequest{
				Email:    identitymocks.DefaultEmail,
				Password: "WrongPass123$",
			},
			expectedErr: errors.New("invalid password"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := identitymocks.NewMockUsersRepository()
			repo.GetUserByEmailFunc = func(ctx context.Context, email *domain.Email) (*domain.User, error) {
				if email.Value() == identitymocks.DefaultEmail {
					return identitymocks.ValidUser(), nil
				}
				return nil, customerrors.NewNotFoundError("user not found")
			}

			uc := newLoginUsecase(
				repo,
				identitymocks.NewMockAccessTokenService(),
				identitymocks.NewMockRefreshSessionRepository(),
				identitymocks.NewMockRefreshTokenGenerator(),
			)

			token, err := uc.Execute(context.Background(), tt.request)
			if err == nil {
				t.Fatalf("expected error %v, got nil", tt.expectedErr)
			}
			if err.Error() != tt.expectedErr.Error() {
				t.Fatalf("expected %v, got: %v", tt.expectedErr, err)
			}
			if token != nil {
				t.Fatalf("expected nil token response on error, got %q", token.AccessToken)
			}
		})
	}
}

func Test_LoginUserUsecase_AccessTokenError(t *testing.T) {
	t.Parallel()

	tokenErr := errors.New("token generation failed")

	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *domain.Email) (*domain.User, error) {
		return identitymocks.ValidUser(), nil
	}

	accessTokens := identitymocks.NewMockAccessTokenService()
	accessTokens.GenerateAccessTokenFunc = func(ctx context.Context, user *identityout.TokenModel) (*identityout.AccessToken, error) {
		return nil, tokenErr
	}

	uc := newLoginUsecase(
		repo,
		accessTokens,
		identitymocks.NewMockRefreshSessionRepository(),
		identitymocks.NewMockRefreshTokenGenerator(),
	)

	_, err := uc.Execute(context.Background(), identityin.LoginUserRequest{
		Email:    identitymocks.DefaultEmail,
		Password: identitymocks.DefaultPassword,
	})
	if !errors.Is(err, tokenErr) {
		t.Fatalf("expected tokenErr, got: %v", err)
	}
}

func Test_LoginUserUsecase_RefreshTokenGeneratorError(t *testing.T) {
	t.Parallel()

	genErr := errors.New("entropy source unavailable")

	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *domain.Email) (*domain.User, error) {
		return identitymocks.ValidUser(), nil
	}

	generator := identitymocks.NewMockRefreshTokenGenerator()
	generator.GenerateFunc = func() (string, error) {
		return "", genErr
	}

	sessionsCreated := false
	sessions := identitymocks.NewMockRefreshSessionRepository()
	sessions.CreateFunc = func(ctx context.Context, userID, tokenHash string) error {
		sessionsCreated = true
		return nil
	}

	uc := newLoginUsecase(
		repo,
		identitymocks.NewMockAccessTokenService(),
		sessions,
		generator,
	)

	_, err := uc.Execute(context.Background(), identityin.LoginUserRequest{
		Email:    identitymocks.DefaultEmail,
		Password: identitymocks.DefaultPassword,
	})
	if !errors.Is(err, genErr) {
		t.Fatalf("expected genErr, got: %v", err)
	}
	if sessionsCreated {
		t.Error("session must not be created when token generation fails")
	}
}

func Test_LoginUserUsecase_SessionCreateError(t *testing.T) {
	t.Parallel()

	storeErr := errors.New("refresh session store unavailable")

	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *domain.Email) (*domain.User, error) {
		return identitymocks.ValidUser(), nil
	}

	sessions := identitymocks.NewMockRefreshSessionRepository()
	sessions.CreateFunc = func(ctx context.Context, userID, tokenHash string) error {
		return storeErr
	}

	uc := newLoginUsecase(
		repo,
		identitymocks.NewMockAccessTokenService(),
		sessions,
		identitymocks.NewMockRefreshTokenGenerator(),
	)

	_, err := uc.Execute(context.Background(), identityin.LoginUserRequest{
		Email:    identitymocks.DefaultEmail,
		Password: identitymocks.DefaultPassword,
	})
	if !errors.Is(err, storeErr) {
		t.Fatalf("expected storeErr, got: %v", err)
	}
}

func Test_LoginUserUsecase_GetUserByEmail_RepoError(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db connection failure")

	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *domain.Email) (*domain.User, error) {
		return nil, dbErr
	}

	uc := newLoginUsecase(
		repo,
		identitymocks.NewMockAccessTokenService(),
		identitymocks.NewMockRefreshSessionRepository(),
		identitymocks.NewMockRefreshTokenGenerator(),
	)

	_, err := uc.Execute(context.Background(), identityin.LoginUserRequest{
		Email:    identitymocks.DefaultEmail,
		Password: identitymocks.DefaultPassword,
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}

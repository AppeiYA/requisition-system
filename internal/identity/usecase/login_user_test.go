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

func Test_LoginUserUsecase_Success(t *testing.T) {
	t.Parallel()

	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *domain.Email) (*domain.User, error) {
		return identitymocks.ValidUser(), nil
	}

	tokenService := identitymocks.NewMockTokenService()
	tokenService.GenerateTokenFunc = func(ctx context.Context, user *identityout.TokenModel) (string, error) {
		if user.ID != identitymocks.DefaultID {
			t.Errorf("expected token model ID %q, got %q", identitymocks.DefaultID, user.ID)
		}
		if user.Email != identitymocks.DefaultEmail {
			t.Errorf("expected token model email %q, got %q", identitymocks.DefaultEmail, user.Email)
		}
		return "test-token", nil
	}

	uc := NewLoginUserUsecase(repo, tokenService, identitymocks.NewMockPasswordHasher())

	token, err := uc.Execute(context.Background(), identityin.LoginUserRequest{
		Email:    identitymocks.DefaultEmail,
		Password: identitymocks.DefaultPassword,
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if token != "test-token" {
		t.Fatalf("expected token 'test-token', got: %q", token)
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

			uc := NewLoginUserUsecase(repo, identitymocks.NewMockTokenService(), identitymocks.NewMockPasswordHasher())

			token, err := uc.Execute(context.Background(), tt.request)
			if err == nil {
				t.Fatalf("expected error %v, got nil", tt.expectedErr)
			}
			if err.Error() != tt.expectedErr.Error() {
				t.Fatalf("expected %v, got: %v", tt.expectedErr, err)
			}
			if token != "" {
				t.Fatalf("expected empty token on error, got %q", token)
			}
		})
	}
}

func Test_LoginUserUsecase_TokenServiceError(t *testing.T) {
	t.Parallel()

	tokenErr := errors.New("token generation failed")

	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *domain.Email) (*domain.User, error) {
		return identitymocks.ValidUser(), nil
	}

	tokenService := identitymocks.NewMockTokenService()
	tokenService.GenerateTokenFunc = func(ctx context.Context, user *identityout.TokenModel) (string, error) {
		return "", tokenErr
	}

	uc := NewLoginUserUsecase(repo, tokenService, identitymocks.NewMockPasswordHasher())

	_, err := uc.Execute(context.Background(), identityin.LoginUserRequest{
		Email:    identitymocks.DefaultEmail,
		Password: identitymocks.DefaultPassword,
	})
	if !errors.Is(err, tokenErr) {
		t.Fatalf("expected tokenErr, got: %v", err)
	}
}

func Test_LoginUserUsecase_GetUserByEmail_RepoError(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db connection failure")

	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *domain.Email) (*domain.User, error) {
		return nil, dbErr
	}

	uc := NewLoginUserUsecase(repo, identitymocks.NewMockTokenService(), identitymocks.NewMockPasswordHasher())

	_, err := uc.Execute(context.Background(), identityin.LoginUserRequest{
		Email:    identitymocks.DefaultEmail,
		Password: identitymocks.DefaultPassword,
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}

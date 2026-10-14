package usecase

import (
	"context"
	"errors"
	"net/http"
	"testing"

	identitymocks "github.com/AppeiYA/requisition-system/internal/identity/mocks"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

func newRefreshUsecase(
	accessTokens identityout.AccessTokenService,
	sessions identityout.RefreshSessionRepository,
	generator identityout.RefreshTokenGenerator,
) *RefreshTokenUsecase {
	return NewRefreshTokenUsecase(accessTokens, sessions, generator)
}

func Test_RefreshTokenUsecase_Rotates(t *testing.T) {
	t.Parallel()

	accessTokens := identitymocks.NewMockAccessTokenService()
	accessTokens.GenerateAccessTokenFunc = func(ctx context.Context, user *identityout.TokenModel) (*identityout.AccessToken, error) {
		if user.ID != identitymocks.DefaultID {
			t.Errorf("expected user ID %q, got %q", identitymocks.DefaultID, user.ID)
		}
		return &identityout.AccessToken{Value: "new-access-token", ExpiresIn: 900}, nil
	}

	generator := identitymocks.NewMockRefreshTokenGenerator()
	generator.GenerateFunc = func() (string, error) {
		return "new-refresh-token", nil
	}

	var gotOldHash, gotNewHash string
	sessions := identitymocks.NewMockRefreshSessionRepository()
	sessions.RotateFunc = func(ctx context.Context, oldTokenHash, newTokenHash string) (*identityout.RefreshSession, error) {
		gotOldHash = oldTokenHash
		gotNewHash = newTokenHash
		return &identityout.RefreshSession{UserID: identitymocks.DefaultID, FamilyID: "family-id"}, nil
	}

	uc := newRefreshUsecase(accessTokens, sessions, generator)

	resp, err := uc.Execute(context.Background(), identityin.RefreshTokenRequest{
		RefreshToken: "presented-refresh-token",
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if gotOldHash != "hash:presented-refresh-token" {
		t.Errorf("expected old token hash %q, got %q", "hash:presented-refresh-token", gotOldHash)
	}
	if gotNewHash != "hash:new-refresh-token" {
		t.Errorf("expected new token hash %q, got %q", "hash:new-refresh-token", gotNewHash)
	}
	if resp.AccessToken != "new-access-token" {
		t.Errorf("expected access token 'new-access-token', got %q", resp.AccessToken)
	}
	if resp.RefreshToken != "new-refresh-token" {
		t.Errorf("expected rotated refresh token 'new-refresh-token', got %q", resp.RefreshToken)
	}
}

func Test_RefreshTokenUsecase_EmptyToken(t *testing.T) {
	t.Parallel()

	uc := newRefreshUsecase(
		identitymocks.NewMockAccessTokenService(),
		identitymocks.NewMockRefreshSessionRepository(),
		identitymocks.NewMockRefreshTokenGenerator(),
	)

	_, err := uc.Execute(context.Background(), identityin.RefreshTokenRequest{})

	assertBadRequest(t, err)
}

func Test_RefreshTokenUsecase_RejectsInvalidTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{"not found", identityout.ErrRefreshSessionNotFound},
		{"revoked", identityout.ErrRefreshSessionRevoked},
		{"reused", identityout.ErrRefreshSessionReused},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sessions := identitymocks.NewMockRefreshSessionRepository()
			sessions.RotateFunc = func(ctx context.Context, oldTokenHash, newTokenHash string) (*identityout.RefreshSession, error) {
				return nil, tt.err
			}

			uc := newRefreshUsecase(
				identitymocks.NewMockAccessTokenService(),
				sessions,
				identitymocks.NewMockRefreshTokenGenerator(),
			)

			_, err := uc.Execute(context.Background(), identityin.RefreshTokenRequest{
				RefreshToken: "presented-refresh-token",
			})
			if !errors.Is(err, tt.err) {
				t.Fatalf("expected %v, got: %v", tt.err, err)
			}
		})
	}
}

func Test_RefreshTokenUsecase_SessionStoreError(t *testing.T) {
	t.Parallel()

	sessions := identitymocks.NewMockRefreshSessionRepository()
	sessions.RotateFunc = func(ctx context.Context, oldTokenHash, newTokenHash string) (*identityout.RefreshSession, error) {
		return nil, identityout.ErrRefreshSessionStore
	}

	uc := newRefreshUsecase(
		identitymocks.NewMockAccessTokenService(),
		sessions,
		identitymocks.NewMockRefreshTokenGenerator(),
	)

	_, err := uc.Execute(context.Background(), identityin.RefreshTokenRequest{
		RefreshToken: "presented-refresh-token",
	})
	if !errors.Is(err, identityout.ErrRefreshSessionStore) {
		t.Fatalf("expected ErrRefreshSessionStore, got: %v", err)
	}
}

func Test_RefreshTokenUsecase_GeneratorError(t *testing.T) {
	t.Parallel()

	genErr := errors.New("entropy source unavailable")

	rotateCalled := false
	sessions := identitymocks.NewMockRefreshSessionRepository()
	sessions.RotateFunc = func(ctx context.Context, oldTokenHash, newTokenHash string) (*identityout.RefreshSession, error) {
		rotateCalled = true
		return &identityout.RefreshSession{UserID: identitymocks.DefaultID}, nil
	}

	generator := identitymocks.NewMockRefreshTokenGenerator()
	generator.GenerateFunc = func() (string, error) {
		return "", genErr
	}

	uc := newRefreshUsecase(identitymocks.NewMockAccessTokenService(), sessions, generator)

	_, err := uc.Execute(context.Background(), identityin.RefreshTokenRequest{
		RefreshToken: "presented-refresh-token",
	})
	if !errors.Is(err, genErr) {
		t.Fatalf("expected genErr, got: %v", err)
	}
	if rotateCalled {
		t.Error("token must not be consumed when a replacement cannot be generated")
	}
}

func Test_RefreshTokenUsecase_AccessTokenError(t *testing.T) {
	t.Parallel()

	accessErr := errors.New("signing failed")

	accessTokens := identitymocks.NewMockAccessTokenService()
	accessTokens.GenerateAccessTokenFunc = func(ctx context.Context, user *identityout.TokenModel) (*identityout.AccessToken, error) {
		return nil, accessErr
	}

	uc := newRefreshUsecase(
		accessTokens,
		identitymocks.NewMockRefreshSessionRepository(),
		identitymocks.NewMockRefreshTokenGenerator(),
	)

	_, err := uc.Execute(context.Background(), identityin.RefreshTokenRequest{
		RefreshToken: "presented-refresh-token",
	})
	if !errors.Is(err, accessErr) {
		t.Fatalf("expected accessErr, got: %v", err)
	}
}

func assertBadRequest(t *testing.T, err error) {
	t.Helper()

	var customErr *customerrors.CustomError
	if !errors.As(err, &customErr) {
		t.Fatalf("expected *customerrors.CustomError, got: %v", err)
	}
	if customErr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, customErr.Code)
	}
}

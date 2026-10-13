package usecase

import (
	"context"
	"errors"
	"testing"

	identitymocks "github.com/AppeiYA/requisition-system/internal/identity/mocks"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

func newLogoutUsecase(
	sessions identityout.RefreshSessionRepository,
	generator identityout.RefreshTokenGenerator,
) *LogoutUsecase {
	return NewLogoutUsecase(sessions, generator)
}

func Test_LogoutUsecase_Revokes(t *testing.T) {
	t.Parallel()

	var revokedHash string
	sessions := identitymocks.NewMockRefreshSessionRepository()
	sessions.RevokeFunc = func(ctx context.Context, tokenHash string) error {
		revokedHash = tokenHash
		return nil
	}

	uc := newLogoutUsecase(sessions, identitymocks.NewMockRefreshTokenGenerator())

	err := uc.Execute(context.Background(), identityin.LogoutRequest{
		RefreshToken: "presented-refresh-token",
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if revokedHash != "hash:presented-refresh-token" {
		t.Errorf("expected revoked hash %q, got %q", "hash:presented-refresh-token", revokedHash)
	}
}

func Test_LogoutUsecase_EmptyToken(t *testing.T) {
	t.Parallel()

	uc := newLogoutUsecase(
		identitymocks.NewMockRefreshSessionRepository(),
		identitymocks.NewMockRefreshTokenGenerator(),
	)

	assertBadRequest(t, uc.Execute(context.Background(), identityin.LogoutRequest{}))
}

func Test_LogoutUsecase_SessionStoreError(t *testing.T) {
	t.Parallel()

	sessions := identitymocks.NewMockRefreshSessionRepository()
	sessions.RevokeFunc = func(ctx context.Context, tokenHash string) error {
		return identityout.ErrRefreshSessionStore
	}

	uc := newLogoutUsecase(sessions, identitymocks.NewMockRefreshTokenGenerator())

	err := uc.Execute(context.Background(), identityin.LogoutRequest{
		RefreshToken: "presented-refresh-token",
	})
	if !errors.Is(err, identityout.ErrRefreshSessionStore) {
		t.Fatalf("expected ErrRefreshSessionStore, got: %v", err)
	}
}

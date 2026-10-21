package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	invitationmocks "github.com/AppeiYA/requisition-system/internal/invitation/mocks"
)

func TestExpireInvitationsUsecase_ExecutesRepositoryTransition(t *testing.T) {
	t.Parallel()

	called := false
	repo := invitationmocks.NewMockInvitationRepository()
	repo.ExpireDueInvitationsFunc = func(ctx context.Context, now time.Time) error {
		called = true
		if now.Location() != time.UTC {
			t.Errorf("expected UTC expiry timestamp, got %v", now.Location())
		}
		return nil
	}

	if err := NewExpireInvitationsUsecase(repo).Execute(context.Background()); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if !called {
		t.Fatal("expected repository expiry transition")
	}
}

func TestExpireInvitationsUsecase_ReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("expire invitations failed")
	repo := invitationmocks.NewMockInvitationRepository()
	repo.ExpireDueInvitationsFunc = func(ctx context.Context, now time.Time) error {
		return wantErr
	}

	if err := NewExpireInvitationsUsecase(repo).Execute(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("expected repository error, got: %v", err)
	}
}

package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	invitationmocks "github.com/AppeiYA/requisition-system/internal/invitation/mocks"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
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

	if err := NewExpireInvitationsUsecase(repo, invitationmocks.NewMockTelemetry()).Execute(context.Background()); err != nil {
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

	if err := NewExpireInvitationsUsecase(repo, invitationmocks.NewMockTelemetry()).Execute(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("expected repository error, got: %v", err)
	}
}

func TestExpireInvitationsUsecase_MeasuresTheSweep(t *testing.T) {
	t.Parallel()

	telemetry := invitationmocks.NewMockTelemetry()
	uc := NewExpireInvitationsUsecase(invitationmocks.NewMockInvitationRepository(), telemetry)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	calls := telemetry.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected exactly one measured operation, got %d", len(calls))
	}
	if calls[0].Operation != invitationout.OperationExpireInvitations {
		t.Errorf("measured operation = %q, want %q", calls[0].Operation, invitationout.OperationExpireInvitations)
	}
}

// TestExpireInvitationsUsecase_TelemetryFailureIsNotTheOutcomeOfTheSweep
// checks the instrumentation cannot turn a sweep that worked into a failed
// one: the scheduler would retry the task for a reason the operation never
// had.
func TestExpireInvitationsUsecase_TelemetryFailureIsNotTheOutcomeOfTheSweep(t *testing.T) {
	t.Parallel()

	telemetry := invitationmocks.NewMockTelemetry()
	telemetry.MeasureFunc = func(ctx context.Context, operation string, fn func(context.Context) error) error {
		_ = fn(ctx)
		return errors.New("telemetry backend down")
	}

	uc := NewExpireInvitationsUsecase(invitationmocks.NewMockInvitationRepository(), telemetry)

	if err := uc.Execute(context.Background()); err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

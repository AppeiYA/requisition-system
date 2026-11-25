package usecase

import (
	"context"
	"testing"
	"time"

	identitymocks "github.com/AppeiYA/requisition-system/internal/identity/mocks"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
)

func Test_CleanupPasswordResetTokensUsecase_UsesRetentionCutoff(t *testing.T) {
	t.Parallel()

	retention := 48 * time.Hour
	repo := identitymocks.NewMockPasswordResetTokenRepository()
	var cutoff time.Time
	repo.DeleteOlderThanFunc = func(ctx context.Context, before time.Time) (int64, error) {
		cutoff = before
		return 3, nil
	}

	uc := NewCleanupPasswordResetTokensUsecase(repo, config.PasswordResetConfig{CleanupRetention: retention})
	count, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if count != 3 {
		t.Fatalf("deleted = %d, want 3", count)
	}

	want := time.Now().UTC().Add(-retention)
	if cutoff.Before(want.Add(-time.Minute)) || cutoff.After(want.Add(time.Minute)) {
		t.Fatalf("cutoff = %v, want about %v", cutoff, want)
	}
}

func Test_SendPasswordResetEmailUsecase_CallsSenderAndMeasures(t *testing.T) {
	t.Parallel()

	sender := identitymocks.NewMockPasswordResetEmailSender()
	telemetry := identitymocks.NewMockTelemetry()
	uc := NewSendPasswordResetEmailUsecase(sender, telemetry)

	expiresAt := time.Now().UTC().Add(30 * time.Minute)
	if err := uc.Execute(context.Background(), identityin.SendPasswordResetEmailCommand{
		To:        "user@example.com",
		Token:     "raw-token",
		ExpiresAt: expiresAt,
	}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if sender.Calls != 1 || sender.To != "user@example.com" || sender.Token != "raw-token" {
		t.Fatalf("unexpected sender call: calls=%d to=%q token=%q", sender.Calls, sender.To, sender.Token)
	}
	if len(telemetry.Operations) != 1 || telemetry.Operations[0] != identityout.OperationSendPasswordResetEmail {
		t.Fatalf("operations = %v", telemetry.Operations)
	}
}

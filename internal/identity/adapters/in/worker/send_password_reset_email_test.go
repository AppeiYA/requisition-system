package identityworkerin

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identitytasks "github.com/AppeiYA/requisition-system/internal/identity/tasks"
	"github.com/hibiken/asynq"
)

type stubSendEmail struct {
	cmd   identityin.SendPasswordResetEmailCommand
	err   error
	calls int
}

func (s *stubSendEmail) Execute(ctx context.Context, cmd identityin.SendPasswordResetEmailCommand) error {
	s.cmd = cmd
	s.calls++
	return s.err
}

func passwordResetTask(t *testing.T, payload identitytasks.PasswordResetEmailPayload) *asynq.Task {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return asynq.NewTask(identitytasks.SendPasswordResetEmail, body)
}

func Test_SendPasswordResetEmailHandler_SkipsBadPayloads(t *testing.T) {
	t.Parallel()

	handler := NewSendPasswordResetEmailHandler(&stubSendEmail{})
	tests := []struct {
		name    string
		payload []byte
	}{
		{name: "invalid json", payload: []byte("not json")},
		{name: "missing recipient", payload: []byte(`{"token":"t","expires_at":"2026-11-25T12:00:00Z"}`)},
		{name: "missing token", payload: []byte(`{"to":"user@example.com","expires_at":"2026-11-25T12:00:00Z"}`)},
		{name: "bad expiry", payload: []byte(`{"to":"user@example.com","token":"t","expires_at":"tomorrow"}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Handle(context.Background(), asynq.NewTask(identitytasks.SendPasswordResetEmail, tt.payload))
			if !errors.Is(err, asynq.SkipRetry) {
				t.Fatalf("error = %v, want SkipRetry", err)
			}
		})
	}
}

func Test_SendPasswordResetEmailHandler_SuccessAndRetry(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2026, time.November, 25, 12, 0, 0, 0, time.UTC)
	payload := identitytasks.PasswordResetEmailPayload{
		ResetID:   "prt-1",
		To:        "user@example.com",
		Token:     "raw-token",
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}

	t.Run("success", func(t *testing.T) {
		stub := &stubSendEmail{}
		handler := NewSendPasswordResetEmailHandler(stub)
		if err := handler.Handle(context.Background(), passwordResetTask(t, payload)); err != nil {
			t.Fatalf("handle: %v", err)
		}
		if stub.calls != 1 || stub.cmd.To != payload.To || stub.cmd.Token != payload.Token || !stub.cmd.ExpiresAt.Equal(expiresAt) {
			t.Fatalf("unexpected command: %+v", stub.cmd)
		}
	})

	t.Run("delivery failure is retried", func(t *testing.T) {
		stub := &stubSendEmail{err: errors.New("provider unavailable")}
		handler := NewSendPasswordResetEmailHandler(stub)
		err := handler.Handle(context.Background(), passwordResetTask(t, payload))
		if err == nil {
			t.Fatalf("expected delivery error")
		}
		if errors.Is(err, asynq.SkipRetry) {
			t.Fatalf("delivery failure must be retried")
		}
	})
}

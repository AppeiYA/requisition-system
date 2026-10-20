package adminworker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	"github.com/AppeiYA/requisition-system/internal/admin/tasks"
	"github.com/hibiken/asynq"
)

type mockSendInvitationEmail struct {
	execute func(ctx context.Context, cmd adminin.SendInvitationEmailCommand) error
	calls   []adminin.SendInvitationEmailCommand
}

func (m *mockSendInvitationEmail) Execute(ctx context.Context, cmd adminin.SendInvitationEmailCommand) error {
	m.calls = append(m.calls, cmd)
	if m.execute != nil {
		return m.execute(ctx, cmd)
	}
	return nil
}

var _ adminin.SendInvitationEmailUseCase = (*mockSendInvitationEmail)(nil)

func invitationEmailTask(t *testing.T, payload tasks.InvitationEmailPayload) *asynq.Task {
	t.Helper()
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	return asynq.NewTask(tasks.SendInvitationEmail, b)
}

func Test_SendMailHandler_Success(t *testing.T) {
	t.Parallel()

	uc := &mockSendInvitationEmail{}
	handler := NewSendMailHandler(uc)

	payload := tasks.InvitationEmailPayload{
		To:             "invitee@example.com",
		InvitedBy:      "usr-1234",
		Token:          "raw-invitation-token",
		ExpiresAt:      "2026-01-01T00:00:00Z",
		DepartmentName: "Finance",
		JobTitle:       "Software Engineer",
	}

	if err := handler.Handle(context.Background(), invitationEmailTask(t, payload)); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if len(uc.calls) != 1 {
		t.Fatalf("expected exactly one use case call, got %d", len(uc.calls))
	}
	got := uc.calls[0]
	if got.To != payload.To || got.InvitedBy != payload.InvitedBy || got.Token != payload.Token || got.ExpiresAt != payload.ExpiresAt {
		t.Errorf("command does not match payload: %+v", got)
	}
	if got.DepartmentName != payload.DepartmentName || got.JobTitle != payload.JobTitle {
		t.Errorf("command does not carry the invite story: %+v", got)
	}
}

func Test_SendMailHandler_InvalidPayloadSkipsRetry(t *testing.T) {
	t.Parallel()

	uc := &mockSendInvitationEmail{}
	handler := NewSendMailHandler(uc)

	task := asynq.NewTask(tasks.SendInvitationEmail, []byte("{not-json"))

	err := handler.Handle(context.Background(), task)
	if err == nil {
		t.Fatal("expected an error for an undecodable payload")
	}
	if !errors.Is(err, asynq.SkipRetry) {
		t.Errorf("expected error to wrap asynq.SkipRetry, got: %v", err)
	}
	if len(uc.calls) != 0 {
		t.Error("use case must not be called for an undecodable payload")
	}
}

func Test_SendMailHandler_UseCaseErrorPropagates(t *testing.T) {
	t.Parallel()

	sendErr := errors.New("send failed")
	uc := &mockSendInvitationEmail{execute: func(ctx context.Context, cmd adminin.SendInvitationEmailCommand) error {
		return sendErr
	}}
	handler := NewSendMailHandler(uc)

	payload := tasks.InvitationEmailPayload{To: "invitee@example.com", Token: "raw-invitation-token"}

	err := handler.Handle(context.Background(), invitationEmailTask(t, payload))
	if !errors.Is(err, sendErr) {
		t.Fatalf("expected sendErr so asynq can retry, got: %v", err)
	}
	if errors.Is(err, asynq.SkipRetry) {
		t.Error("delivery failures must remain retryable")
	}
}

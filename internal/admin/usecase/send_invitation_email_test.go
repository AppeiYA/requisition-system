package usecase

import (
	"context"
	"errors"
	"testing"

	adminmocks "github.com/AppeiYA/requisition-system/internal/admin/mocks"
	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
)

func invitationEmailCmd() adminin.SendInvitationEmailCommand {
	return adminin.SendInvitationEmailCommand{
		To:             adminmocks.DefaultEmail,
		InvitedBy:      adminmocks.DefaultInviterID,
		Token:          adminmocks.DefaultToken,
		ExpiresAt:      "2026-01-01T00:00:00Z",
		DepartmentName: adminmocks.DefaultDepartmentName,
		JobTitle:       adminmocks.DefaultJobTitle,
	}
}

func Test_SendInvitationEmailUsecase_Success(t *testing.T) {
	t.Parallel()

	var gotTo, gotInvitedBy, gotToken, gotExpiresAt, gotDepartmentName, gotJobTitle string
	sender := adminmocks.NewMockEmailSender()
	sender.SendInvitationFunc = func(ctx context.Context, to string, invitedBy string, token string, expiresAt string, departmentName string, jobTitle string) error {
		gotTo = to
		gotInvitedBy = invitedBy
		gotToken = token
		gotExpiresAt = expiresAt
		gotDepartmentName = departmentName
		gotJobTitle = jobTitle
		return nil
	}

	uc := NewSendInvitationEmailUsecase(sender)
	cmd := invitationEmailCmd()

	if err := uc.Execute(context.Background(), cmd); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if gotTo != cmd.To {
		t.Errorf("expected to %q, got %q", cmd.To, gotTo)
	}
	if gotInvitedBy != cmd.InvitedBy {
		t.Errorf("expected invitedBy %q, got %q", cmd.InvitedBy, gotInvitedBy)
	}
	if gotToken != cmd.Token {
		t.Errorf("expected token %q, got %q", cmd.Token, gotToken)
	}
	if gotExpiresAt != cmd.ExpiresAt {
		t.Errorf("expected expiresAt %q, got %q", cmd.ExpiresAt, gotExpiresAt)
	}
	if gotDepartmentName != cmd.DepartmentName {
		t.Errorf("expected departmentName %q, got %q", cmd.DepartmentName, gotDepartmentName)
	}
	if gotJobTitle != cmd.JobTitle {
		t.Errorf("expected jobTitle %q, got %q", cmd.JobTitle, gotJobTitle)
	}
}

func Test_SendInvitationEmailUsecase_SendError(t *testing.T) {
	t.Parallel()

	sendErr := errors.New("send failed")
	sender := adminmocks.NewMockEmailSender()
	sender.SendInvitationFunc = func(ctx context.Context, to string, invitedBy string, token string, expiresAt string, departmentName string, jobTitle string) error {
		return sendErr
	}

	uc := NewSendInvitationEmailUsecase(sender)

	err := uc.Execute(context.Background(), invitationEmailCmd())
	if !errors.Is(err, sendErr) {
		t.Fatalf("expected sendErr, got: %v", err)
	}
}

func Test_SendInvitationEmailUsecase_ContextCancelled(t *testing.T) {
	t.Parallel()

	senderCalled := false
	sender := adminmocks.NewMockEmailSender()
	sender.SendInvitationFunc = func(ctx context.Context, to string, invitedBy string, token string, expiresAt string, departmentName string, jobTitle string) error {
		senderCalled = true
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	uc := NewSendInvitationEmailUsecase(sender)

	err := uc.Execute(ctx, invitationEmailCmd())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
	if senderCalled {
		t.Error("sender must not be called with a cancelled context")
	}
}

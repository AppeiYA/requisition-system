package email

import (
	"context"
	"errors"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/resendlabs/resend-go"
)

type stubEmailsSvc struct {
	got *resend.SendEmailRequest
	err error
}

func (s *stubEmailsSvc) Send(params *resend.SendEmailRequest) (resend.SendEmailResponse, error) {
	s.got = params
	return resend.SendEmailResponse{Id: "email-1"}, s.err
}

func (s *stubEmailsSvc) Get(emailId string) (resend.Email, error) {
	return resend.Email{}, nil
}

func newTestSender(cfg config.EmailConfig, emails resend.EmailsSvc) *Sender {
	return &Sender{emails: emails, cfg: cfg}
}

func Test_Sender_Send(t *testing.T) {
	t.Parallel()

	stub := &stubEmailsSvc{}
	sender := newTestSender(
		config.EmailConfig{APIKey: "re_test", From: "noreply@example.com", FromName: "Requisition System"},
		stub,
	)

	err := sender.Send(context.Background(), "invitee@example.com", "You're invited", "<p>Hello</p>")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if stub.got.From != "Requisition System <noreply@example.com>" {
		t.Errorf("From = %q, want %q", stub.got.From, "Requisition System <noreply@example.com>")
	}
	if len(stub.got.To) != 1 || stub.got.To[0] != "invitee@example.com" {
		t.Errorf("To = %v, want [invitee@example.com]", stub.got.To)
	}
	if stub.got.Subject != "You're invited" {
		t.Errorf("Subject = %q, want %q", stub.got.Subject, "You're invited")
	}
	if stub.got.Html != "<p>Hello</p>" {
		t.Errorf("Html = %q, want %q", stub.got.Html, "<p>Hello</p>")
	}
}

func Test_Sender_Send_WithoutFromName(t *testing.T) {
	t.Parallel()

	stub := &stubEmailsSvc{}
	sender := newTestSender(config.EmailConfig{APIKey: "re_test", From: "noreply@example.com"}, stub)

	if err := sender.Send(context.Background(), "invitee@example.com", "Subject", "Body"); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if stub.got.From != "noreply@example.com" {
		t.Errorf("From = %q, want %q", stub.got.From, "noreply@example.com")
	}
}

func Test_Sender_Send_APIError(t *testing.T) {
	t.Parallel()

	stub := &stubEmailsSvc{err: errors.New("API key is invalid")}
	sender := newTestSender(config.EmailConfig{APIKey: "re_test", From: "noreply@example.com"}, stub)

	err := sender.Send(context.Background(), "invitee@example.com", "Subject", "Body")
	if !errors.Is(err, ErrSendFailed) {
		t.Fatalf("expected ErrSendFailed, got: %v", err)
	}
}

func Test_Sender_Send_ContextCancelled(t *testing.T) {
	t.Parallel()

	stub := &stubEmailsSvc{}
	sender := newTestSender(config.EmailConfig{APIKey: "re_test", From: "noreply@example.com"}, stub)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sender.Send(ctx, "invitee@example.com", "Subject", "Body")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
	if stub.got != nil {
		t.Errorf("expected no API call, got: %+v", stub.got)
	}
}

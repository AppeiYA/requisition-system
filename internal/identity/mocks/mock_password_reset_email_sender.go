package identitymocks

import (
	"context"
	"time"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

type MockPasswordResetEmailSender struct {
	SendFunc func(ctx context.Context, to, token string, expiresAt time.Time) error
	Calls    int
	To       string
	Token    string
}

func NewMockPasswordResetEmailSender() *MockPasswordResetEmailSender {
	return &MockPasswordResetEmailSender{
		SendFunc: func(ctx context.Context, to, token string, expiresAt time.Time) error {
			return nil
		},
	}
}

func (m *MockPasswordResetEmailSender) SendPasswordResetEmail(ctx context.Context, to, token string, expiresAt time.Time) error {
	m.Calls++
	m.To = to
	m.Token = token
	return m.SendFunc(ctx, to, token, expiresAt)
}

var _ identityout.PasswordResetEmailSender = (*MockPasswordResetEmailSender)(nil)

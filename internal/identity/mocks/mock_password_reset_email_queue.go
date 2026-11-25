package identitymocks

import (
	"context"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	identitytasks "github.com/AppeiYA/requisition-system/internal/identity/tasks"
)

type MockPasswordResetEmailQueue struct {
	EnqueueFunc func(ctx context.Context, payload identitytasks.PasswordResetEmailPayload) error
	Payloads    []identitytasks.PasswordResetEmailPayload
}

func NewMockPasswordResetEmailQueue() *MockPasswordResetEmailQueue {
	return &MockPasswordResetEmailQueue{
		EnqueueFunc: func(ctx context.Context, payload identitytasks.PasswordResetEmailPayload) error {
			return nil
		},
	}
}

func (m *MockPasswordResetEmailQueue) EnqueuePasswordResetEmail(ctx context.Context, payload identitytasks.PasswordResetEmailPayload) error {
	if err := m.EnqueueFunc(ctx, payload); err != nil {
		return err
	}
	m.Payloads = append(m.Payloads, payload)
	return nil
}

var _ identityout.PasswordResetEmailQueue = (*MockPasswordResetEmailQueue)(nil)

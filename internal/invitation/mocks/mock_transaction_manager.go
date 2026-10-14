package invitationmocks

import (
	"context"

	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
)

type MockTransactionManager struct {
	WithinTransactionFunc func(ctx context.Context, fn func(context.Context) error) error
}

func NewMockTransactionManager() *MockTransactionManager {
	return &MockTransactionManager{
		WithinTransactionFunc: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	}
}

func (m *MockTransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return m.WithinTransactionFunc(ctx, fn)
}

var _ invitationout.TransactionManager = (*MockTransactionManager)(nil)

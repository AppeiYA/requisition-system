package identitymocks

import (
	"context"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/audit"
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

var _ identityout.TransactionManager = (*MockTransactionManager)(nil)

type MockAuditRecorder struct {
	RecordFunc func(ctx context.Context, entry audit.Entry) error
	Entries    []audit.Entry
}

func NewMockAuditRecorder() *MockAuditRecorder {
	return &MockAuditRecorder{
		RecordFunc: func(ctx context.Context, entry audit.Entry) error {
			return nil
		},
	}
}

func (m *MockAuditRecorder) Record(ctx context.Context, entry audit.Entry) error {
	if err := m.RecordFunc(ctx, entry); err != nil {
		return err
	}
	m.Entries = append(m.Entries, entry)
	return nil
}

var _ identityout.AuditRecorder = (*MockAuditRecorder)(nil)

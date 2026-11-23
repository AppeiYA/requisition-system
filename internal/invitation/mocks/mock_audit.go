package invitationmocks

import (
	"context"

	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/audit"
)

// MockAuditRecorder captures the entries written by the invitation use cases
// so tests can assert the action and the before/after state.
type MockAuditRecorder struct {
	RecordFunc func(ctx context.Context, entry audit.Entry) error

	Entries []audit.Entry
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

var _ invitationout.AuditRecorder = (*MockAuditRecorder)(nil)

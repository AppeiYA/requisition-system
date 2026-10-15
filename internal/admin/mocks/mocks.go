package adminmocks

import (
	"context"
	"time"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

var (
	DefaultInvitationID = "inv-1234"
	DefaultToken        = "raw-invitation-token"
	DefaultInviterID    = "usr-1234"
	DefaultEmail        = "invitee@example.com"
	DefaultFirstName    = "Jane"
	DefaultLastName     = "Doe"
	DefaultPhone        = "+14155551234"
	DefaultStaffID      = "STF-1234"
	DefaultDepartmentID = "dpt-1234"
	DefaultJobTitle     = "Software Engineer"
)

type MockEmailSender struct {
	SendInvitationFunc func(ctx context.Context, to string, invitedBy string, token string, expiresAt string) error
}

func NewMockEmailSender() *MockEmailSender {
	return &MockEmailSender{
		SendInvitationFunc: func(ctx context.Context, to string, invitedBy string, token string, expiresAt string) error {
			return nil
		},
	}
}

func (m *MockEmailSender) SendInvitation(ctx context.Context, to string, invitedBy string, token string, expiresAt string) error {
	return m.SendInvitationFunc(ctx, to, invitedBy, token, expiresAt)
}

var _ adminout.EmailSender = (*MockEmailSender)(nil)

type MockAdminIdentityPort struct {
	CheckUserExistsFunc func(ctx context.Context, email *shareddomain.Email) (bool, error)
}

func NewMockAdminIdentityPort() *MockAdminIdentityPort {
	return &MockAdminIdentityPort{
		CheckUserExistsFunc: func(ctx context.Context, email *shareddomain.Email) (bool, error) {
			return false, nil
		},
	}
}

func (m *MockAdminIdentityPort) CheckUserExists(ctx context.Context, email *shareddomain.Email) (bool, error) {
	return m.CheckUserExistsFunc(ctx, email)
}

var _ adminout.AdminIdentityPort = (*MockAdminIdentityPort)(nil)

type MockAdminDepartmentPort struct {
	CheckDepartmentExistsFunc func(ctx context.Context, departmentID string) (bool, error)
}

func NewMockAdminDepartmentPort() *MockAdminDepartmentPort {
	return &MockAdminDepartmentPort{
		CheckDepartmentExistsFunc: func(ctx context.Context, departmentID string) (bool, error) {
			return true, nil
		},
	}
}

func (m *MockAdminDepartmentPort) CheckDepartmentExists(ctx context.Context, departmentID string) (bool, error) {
	return m.CheckDepartmentExistsFunc(ctx, departmentID)
}

var _ adminout.AdminDepartmentPort = (*MockAdminDepartmentPort)(nil)

type MockAdminInvitationPort struct {
	CheckActiveInvitationExistsFunc func(ctx context.Context, email *shareddomain.Email) (bool, error)
	SaveInvitationFunc              func(ctx context.Context, params adminout.SaveInvitationParams) (*adminout.SavedInvitation, error)
}

func NewMockAdminInvitationPort() *MockAdminInvitationPort {
	return &MockAdminInvitationPort{
		CheckActiveInvitationExistsFunc: func(ctx context.Context, email *shareddomain.Email) (bool, error) {
			return false, nil
		},
		SaveInvitationFunc: func(ctx context.Context, params adminout.SaveInvitationParams) (*adminout.SavedInvitation, error) {
			return &adminout.SavedInvitation{
				ID:        DefaultInvitationID,
				Token:     DefaultToken,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}, nil
		},
	}
}

func (m *MockAdminInvitationPort) CheckActiveInvitationExists(ctx context.Context, email *shareddomain.Email) (bool, error) {
	return m.CheckActiveInvitationExistsFunc(ctx, email)
}

func (m *MockAdminInvitationPort) SaveInvitation(ctx context.Context, params adminout.SaveInvitationParams) (*adminout.SavedInvitation, error) {
	return m.SaveInvitationFunc(ctx, params)
}

var _ adminout.AdminInvitationPort = (*MockAdminInvitationPort)(nil)

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

var _ adminout.TransactionManager = (*MockTransactionManager)(nil)

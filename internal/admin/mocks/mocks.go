package adminmocks

import (
	"context"
	"time"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	"github.com/AppeiYA/requisition-system/internal/admin/tasks"
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
	// DefaultDepartmentName mirrors the name the department module returns
	// for DefaultDepartmentID, so queued payloads can be asserted end to end.
	DefaultDepartmentName = "Finance"
)

type MockEmailSender struct {
	SendInvitationFunc func(ctx context.Context, to string, invitedBy string, token string, expiresAt string, departmentName string, jobTitle string) error
}

func NewMockEmailSender() *MockEmailSender {
	return &MockEmailSender{
		SendInvitationFunc: func(ctx context.Context, to string, invitedBy string, token string, expiresAt string, departmentName string, jobTitle string) error {
			return nil
		},
	}
}

func (m *MockEmailSender) SendInvitation(ctx context.Context, to string, invitedBy string, token string, expiresAt string, departmentName string, jobTitle string) error {
	return m.SendInvitationFunc(ctx, to, invitedBy, token, expiresAt, departmentName, jobTitle)
}

var _ adminout.EmailSender = (*MockEmailSender)(nil)

type MockTaskQueue struct {
	EnqueueInvitationEmailFunc func(ctx context.Context, payload tasks.InvitationEmailPayload) error
}

func NewMockTaskQueue() *MockTaskQueue {
	return &MockTaskQueue{
		EnqueueInvitationEmailFunc: func(ctx context.Context, payload tasks.InvitationEmailPayload) error {
			return nil
		},
	}
}

func (m *MockTaskQueue) EnqueueInvitationEmail(ctx context.Context, payload tasks.InvitationEmailPayload) error {
	return m.EnqueueInvitationEmailFunc(ctx, payload)
}

var _ adminout.TaskQueue = (*MockTaskQueue)(nil)

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
	GetDepartmentNameFunc     func(ctx context.Context, departmentID string) (string, error)
	GetDepartmentByIDFunc     func(ctx context.Context, departmentID string) (*adminout.DepartmentResponse, error)
}

func NewMockAdminDepartmentPort() *MockAdminDepartmentPort {
	return &MockAdminDepartmentPort{
		CheckDepartmentExistsFunc: func(ctx context.Context, departmentID string) (bool, error) {
			return true, nil
		},
		GetDepartmentNameFunc: func(ctx context.Context, departmentID string) (string, error) {
			return DefaultDepartmentName, nil
		},
		GetDepartmentByIDFunc: func(ctx context.Context, departmentID string) (*adminout.DepartmentResponse, error) {
			return &adminout.DepartmentResponse{
				ID:   departmentID,
				Name: DefaultDepartmentName,
			}, nil
		},
	}
}

func (m *MockAdminDepartmentPort) CheckDepartmentExists(ctx context.Context, departmentID string) (bool, error) {
	return m.CheckDepartmentExistsFunc(ctx, departmentID)
}

func (m *MockAdminDepartmentPort) GetDepartmentName(ctx context.Context, departmentID string) (string, error) {
	return m.GetDepartmentNameFunc(ctx, departmentID)
}

func (m *MockAdminDepartmentPort) GetDepartmentByID(ctx context.Context, departmentID string) (*adminout.DepartmentResponse, error) {
	return m.GetDepartmentByIDFunc(ctx, departmentID)
}

var _ adminout.AdminDepartmentPort = (*MockAdminDepartmentPort)(nil)

type MockAdminInvitationPort struct {
	SaveInvitationFunc func(ctx context.Context, params adminout.SaveInvitationParams) (*adminout.SavedInvitation, error)
}

func NewMockAdminInvitationPort() *MockAdminInvitationPort {
	return &MockAdminInvitationPort{
		SaveInvitationFunc: func(ctx context.Context, params adminout.SaveInvitationParams) (*adminout.SavedInvitation, error) {
			return &adminout.SavedInvitation{
				ID:        DefaultInvitationID,
				Token:     DefaultToken,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}, nil
		},
	}
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

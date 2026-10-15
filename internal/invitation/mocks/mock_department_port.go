package invitationmocks

import (
	"context"

	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
)

type MockInvitationDepartmentPort struct {
	CreateMembershipFunc func(ctx context.Context, req *invitationout.CreateMembershipRequest) error
}

func NewMockInvitationDepartmentPort() *MockInvitationDepartmentPort {
	return &MockInvitationDepartmentPort{
		CreateMembershipFunc: func(ctx context.Context, req *invitationout.CreateMembershipRequest) error {
			return nil
		},
	}
}

func (m *MockInvitationDepartmentPort) CreateMembership(ctx context.Context, req *invitationout.CreateMembershipRequest) error {
	return m.CreateMembershipFunc(ctx, req)
}

var _ invitationout.InvitationDepartmentPort = (*MockInvitationDepartmentPort)(nil)

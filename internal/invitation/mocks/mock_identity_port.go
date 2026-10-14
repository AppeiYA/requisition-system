package invitationmocks

import (
	"context"

	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
)

type MockInvitationIdentityPort struct {
	RegisterUserFunc func(ctx context.Context, req *invitationout.CreateIdentityRequest) error
}

func NewMockInvitationIdentityPort() *MockInvitationIdentityPort {
	return &MockInvitationIdentityPort{
		RegisterUserFunc: func(ctx context.Context, req *invitationout.CreateIdentityRequest) error {
			return nil
		},
	}
}

func (m *MockInvitationIdentityPort) RegisterUser(ctx context.Context, req *invitationout.CreateIdentityRequest) error {
	return m.RegisterUserFunc(ctx, req)
}

var _ invitationout.InvitationIdentityPort = (*MockInvitationIdentityPort)(nil)

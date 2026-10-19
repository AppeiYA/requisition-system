package invitationmocks

import (
	"context"

	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
)

var DefaultRegisteredUserID = "usr-5678"

type MockInvitationIdentityPort struct {
	RegisterUserFunc func(ctx context.Context, req *invitationout.CreateIdentityRequest) (string, error)
}

func NewMockInvitationIdentityPort() *MockInvitationIdentityPort {
	return &MockInvitationIdentityPort{
		RegisterUserFunc: func(ctx context.Context, req *invitationout.CreateIdentityRequest) (string, error) {
			return DefaultRegisteredUserID, nil
		},
	}
}

func (m *MockInvitationIdentityPort) RegisterUser(ctx context.Context, req *invitationout.CreateIdentityRequest) (string, error) {
	return m.RegisterUserFunc(ctx, req)
}

var _ invitationout.InvitationIdentityPort = (*MockInvitationIdentityPort)(nil)

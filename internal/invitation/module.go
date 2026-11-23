package invitation

import (
	invitationhttp "github.com/AppeiYA/requisition-system/internal/invitation/adapters/in/http"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
)

type InvitationModule struct {
	invitationPortIn       *invitationin.InvitationPortIn
	adminInvitationsPortIn *invitationin.AdminInvitationsPortIn
	invitationHandler      *invitationhttp.InvitationHandler
}

func NewInvitationModule(
	invitationPortIn *invitationin.InvitationPortIn,
	adminInvitationsPortIn *invitationin.AdminInvitationsPortIn,
	invitationHandler *invitationhttp.InvitationHandler,
) *InvitationModule {
	return &InvitationModule{
		invitationPortIn:       invitationPortIn,
		adminInvitationsPortIn: adminInvitationsPortIn,
		invitationHandler:      invitationHandler,
	}
}

func (m *InvitationModule) PortIn() *invitationin.InvitationPortIn {
	return m.invitationPortIn
}

func (m *InvitationModule) Handler() *invitationhttp.InvitationHandler {
	return m.invitationHandler
}

// AdminInvitations exposes the admin-facing use cases to the admin module's
// external adapter.
func (m *InvitationModule) AdminInvitations() *invitationin.AdminInvitationsPortIn {
	return m.adminInvitationsPortIn
}

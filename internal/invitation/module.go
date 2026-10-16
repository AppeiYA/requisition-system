package invitation

import (
	invitationhttp "github.com/AppeiYA/requisition-system/internal/invitation/adapters/in/http"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
)

type InvitationModule struct {
	invitationPortIn *invitationin.InvitationPortIn
	invitationHandler *invitationhttp.InvitationHandler
}

func NewInvitationModule(
	invitationPortIn *invitationin.InvitationPortIn,
	invitationHandler *invitationhttp.InvitationHandler,
) *InvitationModule {
	return &InvitationModule{
		invitationPortIn: invitationPortIn,
		invitationHandler: invitationHandler,
	}
}

func (m *InvitationModule) PortIn() *invitationin.InvitationPortIn {
	return m.invitationPortIn
}

func (m *InvitationModule) Handler() *invitationhttp.InvitationHandler {
	return m.invitationHandler
}
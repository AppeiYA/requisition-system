package invitationexternal

import (
	"context"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
)

var _ invitationout.InvitationIdentityPort = (*InviteIdentityAdapter)(nil)

type InviteIdentityAdapter struct {
	registerUC identityin.RegisterUserUseCaseInt
}

func NewInviteIdentityAdapter(registerUC identityin.RegisterUserUseCaseInt) *InviteIdentityAdapter {
	return &InviteIdentityAdapter{
		registerUC: registerUC,
	}
}

func (a *InviteIdentityAdapter) RegisterUser(ctx context.Context, dto *invitationout.CreateIdentityRequest) error {
	req := identityin.RegisterUserDTO{
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		Email:     dto.Email,
		Password:  dto.Password,
	}

	return a.registerUC.Execute(ctx, req)
}

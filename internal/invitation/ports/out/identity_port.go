package invitationout

import "context"

type CreateIdentityRequest struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
}

type InvitationIdentityPort interface {
	RegisterUser(ctx context.Context, req *CreateIdentityRequest) error
}

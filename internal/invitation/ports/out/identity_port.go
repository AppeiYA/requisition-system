package invitationout

import "context"

type CreateIdentityRequest struct {
	FirstName string
	LastName  string
	Phone     string
	StaffID   string
	Email     string
	Password  string
}

type InvitationIdentityPort interface {
	// RegisterUser creates the invitee's user account and returns its ID so
	// the caller can link the department membership in the same transaction.
	RegisterUser(ctx context.Context, req *CreateIdentityRequest) (string, error)
}

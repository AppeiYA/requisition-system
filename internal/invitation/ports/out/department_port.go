package invitationout

import "context"

// CreateMembershipRequest links a newly registered user to the department they
// were invited to.
type CreateMembershipRequest struct {
	UserID       string
	DepartmentID string
	JobTitle     string
}

// InvitationDepartmentPort is the invitation module's gateway to the
// department module. It creates the membership once the user exists.
type InvitationDepartmentPort interface {
	CreateMembership(ctx context.Context, req *CreateMembershipRequest) error
}

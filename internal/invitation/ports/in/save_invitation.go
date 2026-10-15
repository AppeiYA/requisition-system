package invitationin

import (
	"context"
	"time"
)

// SaveInvitationCommand carries the data needed to create an invitation. The
// name and phone fields are the admin-provided prefill values. DepartmentID
// and JobTitle describe the membership the invitee will receive on acceptance.
type SaveInvitationCommand struct {
	Email        string
	FirstName    string
	LastName     string
	Phone        string
	StaffID      string
	DepartmentID string
	JobTitle     string
	InviterID    string
}

type SaveInvitationResponse struct {
	ID        string
	Token     string
	ExpiresAt time.Time
}

type SaveInvitationUsecaseInt interface {
	Execute(ctx context.Context, cmd SaveInvitationCommand) (*SaveInvitationResponse, error)
}

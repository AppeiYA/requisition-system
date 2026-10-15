package invitationin

import (
	"context"
)

// ValidateInvitationResult describes whether an invitation token is valid and,
// when it is, the prefilled registration details for the invitee.
type ValidateInvitationResult struct {
	Valid        bool
	Email        string
	FirstName    string
	LastName     string
	Phone        string
	StaffID      string
	DepartmentID string
	JobTitle     string
}

type ValidateInvitationUseCaseInt interface {
	Execute(ctx context.Context, token string) (*ValidateInvitationResult, error)
}

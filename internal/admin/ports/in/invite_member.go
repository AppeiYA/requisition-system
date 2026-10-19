package adminin

import "context"

type InviteMemberCommand struct {
	InviterID    string
	Email        string
	FirstName    string
	LastName     string
	Phone        string
	StaffID      string
	DepartmentID string
	JobTitle     string
}

type InviteMemberResult struct {
	InvitationID string
}

type InviteMemberUseCase interface {
	Execute(ctx context.Context, cmd InviteMemberCommand) (*InviteMemberResult, error)
}

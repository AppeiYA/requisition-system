package adminin

import "context"

type InviteMemberCommand struct {
	InviterID string
	Email     string
}

type InviteMemberResult struct {
	InvitationID string
}

type InviteMemberUseCase interface {
	Execute(ctx context.Context, cmd InviteMemberCommand) (*InviteMemberResult, error)
}

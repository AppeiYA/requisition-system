package invitationin

import "context"

type AcceptInvitationCommand struct {
	Token     string
	FirstName string
	LastName  string
	Password  string
}

type AcceptInvitationUseCaseInt interface {
	Execute(ctx context.Context, cmd AcceptInvitationCommand) error
}

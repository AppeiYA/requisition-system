package invitationin

import "context"

// AcceptInvitationCommand carries the invitee-supplied registration data. The
// first/last name are fixed by the invitation's prefill (the invitee cannot
// change them). Phone is optional: when omitted, the prefilled value is used.
type AcceptInvitationCommand struct {
	Token    string
	Phone    string
	Password string
}

type AcceptInvitationUseCaseInt interface {
	Execute(ctx context.Context, cmd AcceptInvitationCommand) error
}

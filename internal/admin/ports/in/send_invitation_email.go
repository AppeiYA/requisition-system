package adminin

import "context"

// SendInvitationEmailCommand carries everything needed to deliver one
// invitation email.
type SendInvitationEmailCommand struct {
	To        string
	InvitedBy string
	Token     string
	ExpiresAt string
	// DepartmentName and JobTitle are rendered into the invitee-facing
	// story line of the mail.
	DepartmentName string
	JobTitle       string
}

// SendInvitationEmailUseCase delivers an invitation email. It is driven by
// the worker process through the queued task, not by the HTTP layer.
type SendInvitationEmailUseCase interface {
	Execute(ctx context.Context, cmd SendInvitationEmailCommand) error
}

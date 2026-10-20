package usecase

import (
	"context"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
)

// SendInvitationEmailUsecase delivers an invitation email through the
// EmailSender port. It runs in the worker process, driven by the queued
// task, so a slow or failing mail provider never blocks the invite request.
type SendInvitationEmailUsecase struct {
	emailSender adminout.EmailSender
}

func NewSendInvitationEmailUsecase(emailSender adminout.EmailSender) *SendInvitationEmailUsecase {
	return &SendInvitationEmailUsecase{
		emailSender: emailSender,
	}
}

func (u *SendInvitationEmailUsecase) Execute(ctx context.Context, cmd adminin.SendInvitationEmailCommand) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return u.emailSender.SendInvitation(ctx, cmd.To, cmd.InvitedBy, cmd.Token, cmd.ExpiresAt, cmd.DepartmentName, cmd.JobTitle)
}

package usecase

import (
	"context"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
)

type SendInvitationEmailUsecase struct {
	emailSender adminout.EmailSender
	telemetry   adminout.Telemetry
}

func NewSendInvitationEmailUsecase(emailSender adminout.EmailSender, telemetry adminout.Telemetry) *SendInvitationEmailUsecase {
	return &SendInvitationEmailUsecase{
		emailSender: emailSender,
		telemetry:   telemetry,
	}
}

func (u *SendInvitationEmailUsecase) Execute(ctx context.Context, cmd adminin.SendInvitationEmailCommand) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var operationErr error
	_ = u.telemetry.Measure(ctx, adminout.OperationSendInvitationEmail, func(ctx context.Context) error {
		operationErr = u.emailSender.SendInvitation(ctx, cmd.To, cmd.InvitedBy, cmd.Token, cmd.ExpiresAt, cmd.DepartmentName, cmd.JobTitle)
		return operationErr
	})

	return operationErr
}

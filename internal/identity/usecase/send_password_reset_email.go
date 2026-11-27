package usecase

import (
	"context"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

// SendPasswordResetEmailUsecase is the worker-side operation that delivers a
// reset link. It is measured like the other business operations of identity.
type SendPasswordResetEmailUsecase struct {
	sender    identityout.PasswordResetEmailSender
	telemetry identityout.Telemetry
}

func NewSendPasswordResetEmailUsecase(
	sender identityout.PasswordResetEmailSender,
	telemetry identityout.Telemetry,
) *SendPasswordResetEmailUsecase {
	return &SendPasswordResetEmailUsecase{
		sender:    sender,
		telemetry: telemetry,
	}
}

func (u *SendPasswordResetEmailUsecase) Execute(ctx context.Context, cmd identityin.SendPasswordResetEmailCommand) error {
	return u.telemetry.Measure(ctx, identityout.OperationSendPasswordResetEmail, func(ctx context.Context) error {
		return u.sender.SendPasswordResetEmail(ctx, cmd.To, cmd.Token, cmd.ExpiresAt)
	})
}

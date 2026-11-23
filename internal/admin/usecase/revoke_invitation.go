package usecase

import (
	"context"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
)

type RevokeInvitationUsecase struct {
	invitations adminout.AdminInvitationManagementPort
	telemetry   adminout.Telemetry
}

func NewRevokeInvitationUsecase(invitations adminout.AdminInvitationManagementPort, telemetry adminout.Telemetry) *RevokeInvitationUsecase {
	return &RevokeInvitationUsecase{invitations: invitations, telemetry: telemetry}
}

func (uc *RevokeInvitationUsecase) Execute(ctx context.Context, cmd adminin.RevokeInvitationCommand) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var operationErr error
	_ = uc.telemetry.Measure(ctx, adminout.OperationRevokeInvitation, func(ctx context.Context) error {
		operationErr = uc.invitations.RevokeInvitation(ctx, adminout.RevokeInvitationParams{
			InvitationID: cmd.InvitationID,
			ActorID:      cmd.ActorID,
			Reason:       cmd.Reason,
		})
		return operationErr
	})
	return operationErr
}

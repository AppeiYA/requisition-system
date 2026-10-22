package usecase

import (
	"context"
	"time"

	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
)

type ExpireInvitationsUsecase struct {
	invitationRepo invitationout.InvitationRepository
	telemetry      invitationout.Telemetry
}

func NewExpireInvitationsUsecase(invitationRepo invitationout.InvitationRepository, telemetry invitationout.Telemetry) *ExpireInvitationsUsecase {
	return &ExpireInvitationsUsecase{
		invitationRepo: invitationRepo,
		telemetry:      telemetry,
	}
}

func (uc *ExpireInvitationsUsecase) Execute(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var operationErr error
	_ = uc.telemetry.Measure(ctx, invitationout.OperationExpireInvitations, func(ctx context.Context) error {
		operationErr = uc.invitationRepo.ExpireDueInvitations(ctx, time.Now().UTC())
		return operationErr
	})

	return operationErr
}

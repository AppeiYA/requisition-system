package usecase

import (
	"context"
	"time"

	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
)

type ExpireInvitationsUsecase struct {
	invitationRepo invitationout.InvitationRepository
}

func NewExpireInvitationsUsecase(invitationRepo invitationout.InvitationRepository) *ExpireInvitationsUsecase {
	return &ExpireInvitationsUsecase{invitationRepo: invitationRepo}
}

func (uc *ExpireInvitationsUsecase) Execute(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return uc.invitationRepo.ExpireDueInvitations(ctx, time.Now().UTC())
}

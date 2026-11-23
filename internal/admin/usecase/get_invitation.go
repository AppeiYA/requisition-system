package usecase

import (
	"context"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
)

type GetInvitationUsecase struct {
	invitations adminout.AdminInvitationManagementPort
}

func NewGetInvitationUsecase(invitations adminout.AdminInvitationManagementPort) *GetInvitationUsecase {
	return &GetInvitationUsecase{invitations: invitations}
}

func (uc *GetInvitationUsecase) Execute(ctx context.Context, invitationID string) (*adminout.InvitationSummary, error) {
	return uc.invitations.GetInvitation(ctx, invitationID)
}

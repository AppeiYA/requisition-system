package usecase

import (
	"context"

	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
)

type AdminGetInvitationUsecase struct {
	repo invitationout.InvitationRepository
}

func NewAdminGetInvitationUsecase(repo invitationout.InvitationRepository) *AdminGetInvitationUsecase {
	return &AdminGetInvitationUsecase{repo: repo}
}

func (uc *AdminGetInvitationUsecase) Execute(ctx context.Context, invitationID string) (*invitationdomain.InvitationSummary, error) {
	if invitationID == "" {
		return nil, ErrInvitationIDRequired
	}
	return uc.repo.GetInvitationSummary(ctx, invitationID)
}

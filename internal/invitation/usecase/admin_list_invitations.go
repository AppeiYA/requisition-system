package usecase

import (
	"context"

	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

type AdminListInvitationsUsecase struct {
	repo invitationout.InvitationRepository
}

func NewAdminListInvitationsUsecase(repo invitationout.InvitationRepository) *AdminListInvitationsUsecase {
	return &AdminListInvitationsUsecase{repo: repo}
}

func (uc *AdminListInvitationsUsecase) Execute(ctx context.Context, filter invitationin.AdminInvitationListFilter) (*invitationin.AdminListInvitationsResult, error) {
	if filter.Status != "" {
		if _, err := invitationdomain.NewInvitationStatus(filter.Status); err != nil {
			return nil, err
		}
	}

	limit, offset := shareddomain.Clamp(filter.Limit, filter.Offset)
	domainFilter := invitationdomain.InvitationListFilter{
		Search:       filter.Search,
		Status:       filter.Status,
		DepartmentID: filter.DepartmentID,
		InvitedBy:    filter.InvitedBy,
		From:         filter.From,
		To:           filter.To,
		Limit:        limit,
		Offset:       offset,
	}

	invitations, err := uc.repo.ListInvitations(ctx, domainFilter)
	if err != nil {
		return nil, err
	}
	total, err := uc.repo.CountInvitations(ctx, domainFilter)
	if err != nil {
		return nil, err
	}
	return &invitationin.AdminListInvitationsResult{Invitations: invitations, Total: total}, nil
}

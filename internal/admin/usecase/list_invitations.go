package usecase

import (
	"context"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

type ListInvitationsUsecase struct {
	invitations adminout.AdminInvitationManagementPort
}

func NewListInvitationsUsecase(invitations adminout.AdminInvitationManagementPort) *ListInvitationsUsecase {
	return &ListInvitationsUsecase{invitations: invitations}
}

func (uc *ListInvitationsUsecase) Execute(ctx context.Context, filter adminin.InvitationListFilter) (*adminin.ListInvitationsResult, error) {
	limit, offset := shareddomain.Clamp(filter.Limit, filter.Offset)
	invitations, total, err := uc.invitations.ListInvitations(ctx, adminout.InvitationListFilter{
		Search:       filter.Search,
		Status:       filter.Status,
		DepartmentID: filter.DepartmentID,
		InvitedBy:    filter.InvitedBy,
		From:         filter.From,
		To:           filter.To,
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		return nil, err
	}
	return &adminin.ListInvitationsResult{Invitations: invitations, Total: total}, nil
}

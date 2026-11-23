package adminexternal

import (
	"context"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
)

var _ adminout.AdminInvitationManagementPort = (*AdminInvitationManagementAdapter)(nil)

// AdminInvitationManagementAdapter adapts the invitation module's admin use
// cases to the admin module's outbound port. It maps the invitation read model
// into a tag-free admin DTO, so no invitation domain type crosses this boundary.
type AdminInvitationManagementAdapter struct {
	invitations *invitationin.AdminInvitationsPortIn
}

func NewAdminInvitationManagementAdapter(invitations *invitationin.AdminInvitationsPortIn) *AdminInvitationManagementAdapter {
	return &AdminInvitationManagementAdapter{invitations: invitations}
}

func (a *AdminInvitationManagementAdapter) ListInvitations(ctx context.Context, filter adminout.InvitationListFilter) ([]adminout.InvitationSummary, int, error) {
	result, err := a.invitations.ListInvitations.Execute(ctx, invitationin.AdminInvitationListFilter{
		Search:       filter.Search,
		Status:       filter.Status,
		DepartmentID: filter.DepartmentID,
		InvitedBy:    filter.InvitedBy,
		From:         filter.From,
		To:           filter.To,
		Limit:        filter.Limit,
		Offset:       filter.Offset,
	})
	if err != nil {
		return nil, 0, err
	}

	summaries := make([]adminout.InvitationSummary, len(result.Invitations))
	for i, invitation := range result.Invitations {
		summaries[i] = toAdminInvitationSummary(invitation)
	}
	return summaries, result.Total, nil
}

func (a *AdminInvitationManagementAdapter) GetInvitation(ctx context.Context, invitationID string) (*adminout.InvitationSummary, error) {
	summary, err := a.invitations.GetInvitation.Execute(ctx, invitationID)
	if err != nil {
		return nil, err
	}
	result := toAdminInvitationSummary(*summary)
	return &result, nil
}

func toAdminInvitationSummary(summary invitationdomain.InvitationSummary) adminout.InvitationSummary {
	return adminout.InvitationSummary{
		ID:               summary.ID,
		FirstName:        summary.FirstName,
		LastName:         summary.LastName,
		Email:            summary.Email,
		DepartmentID:     summary.DepartmentID,
		DepartmentName:   summary.DepartmentName,
		JobTitle:         summary.JobTitle,
		Status:           summary.Status,
		InvitedBy:        summary.InvitedBy,
		InviterName:      summary.InviterName,
		ExpiresAt:        summary.ExpiresAt,
		AcceptedAt:       summary.AcceptedAt,
		RevokedAt:        summary.RevokedAt,
		DeliveredAt:      summary.DeliveredAt,
		DeliveryFailedAt: summary.DeliveryFailedAt,
		CreatedAt:        summary.CreatedAt,
	}
}

func (a *AdminInvitationManagementAdapter) ResendInvitation(ctx context.Context, params adminout.ResendInvitationParams) (*adminout.ResendInvitationResult, error) {
	result, err := a.invitations.ResendInvitation.Execute(ctx, invitationin.ResendInvitationCommand{
		InvitationID: params.InvitationID,
		ActorID:      params.ActorID,
	})
	if err != nil {
		return nil, err
	}
	return &adminout.ResendInvitationResult{
		ID:           result.ID,
		Token:        result.Token,
		Email:        result.Email,
		FirstName:    result.FirstName,
		LastName:     result.LastName,
		InvitedBy:    result.InvitedBy,
		DepartmentID: result.DepartmentID,
		JobTitle:     result.JobTitle,
		ExpiresAt:    result.ExpiresAt,
	}, nil
}

func (a *AdminInvitationManagementAdapter) RevokeInvitation(ctx context.Context, params adminout.RevokeInvitationParams) error {
	return a.invitations.RevokeInvitation.Execute(ctx, invitationin.RevokeInvitationCommand{
		InvitationID: params.InvitationID,
		ActorID:      params.ActorID,
		Reason:       params.Reason,
	})
}

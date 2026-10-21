package adminexternal

import (
	"context"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
)

var _ adminout.AdminInvitationPort = (*AdminInvitationAdapter)(nil)

// AdminInvitationAdapter adapts the invitation module to the admin module's
// outbound port. It delegates persistence to the invitation save use case and
// reads pending invitations through the invitation repository.
type AdminInvitationAdapter struct {
	saveInvitation invitationin.SaveInvitationUsecaseInt
}

func NewAdminInvitationAdapter(
	saveInvitation invitationin.SaveInvitationUsecaseInt,
) *AdminInvitationAdapter {
	return &AdminInvitationAdapter{
		saveInvitation: saveInvitation,
	}
}

func (a *AdminInvitationAdapter) SaveInvitation(ctx context.Context, params adminout.SaveInvitationParams) (*adminout.SavedInvitation, error) {
	res, err := a.saveInvitation.Execute(ctx, invitationin.SaveInvitationCommand{
		Email:        params.Email.Value(),
		FirstName:    params.FirstName,
		LastName:     params.LastName,
		Phone:        params.Phone,
		StaffID:      params.StaffID,
		DepartmentID: params.DepartmentID,
		JobTitle:     params.JobTitle,
		InviterID:    params.InviterID,
	})
	if err != nil {
		return nil, err
	}

	return &adminout.SavedInvitation{
		ID:        res.ID,
		Token:     res.Token,
		ExpiresAt: res.ExpiresAt,
	}, nil
}

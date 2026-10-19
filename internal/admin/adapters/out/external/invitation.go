package adminexternal

import (
	"context"
	"errors"
	"net/http"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var _ adminout.AdminInvitationPort = (*AdminInvitationAdapter)(nil)

// AdminInvitationAdapter adapts the invitation module to the admin module's
// outbound port. It delegates persistence to the invitation save use case and
// reads pending invitations through the invitation repository.
type AdminInvitationAdapter struct {
	saveInvitation invitationin.SaveInvitationUsecaseInt
	getInvitation invitationin.GetInvitationUsecaseInt
}

func NewAdminInvitationAdapter(
	saveInvitation invitationin.SaveInvitationUsecaseInt,
	getInvitation invitationin.GetInvitationUsecaseInt,
) *AdminInvitationAdapter {
	return &AdminInvitationAdapter{
		saveInvitation: saveInvitation,
		getInvitation: getInvitation,
	}
}

func (a *AdminInvitationAdapter) CheckActiveInvitationExists(ctx context.Context, email *shareddomain.Email) (bool, error) {
	_, err := a.getInvitation.ByEmail(ctx, email.Value())
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
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

func isNotFound(err error) bool {
	var custErr *customerrors.CustomError
	return errors.As(err, &custErr) && custErr.Code == http.StatusNotFound
}

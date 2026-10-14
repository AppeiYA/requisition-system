package adminexternal

import (
	"context"
	"errors"
	"net/http"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var _ adminout.AdminInvitationPort = (*AdminInvitationAdapter)(nil)

// AdminInvitationAdapter adapts the invitation module to the admin module's
// outbound port. It delegates persistence to the invitation save use case and
// reads pending invitations through the invitation repository.
type AdminInvitationAdapter struct {
	saveInvitation invitationin.SaveInvitationUsecaseInt
	invitationRepo invitationout.InvitationRepository
}

func NewAdminInvitationAdapter(
	saveInvitation invitationin.SaveInvitationUsecaseInt,
	invitationRepo invitationout.InvitationRepository,
) *AdminInvitationAdapter {
	return &AdminInvitationAdapter{
		saveInvitation: saveInvitation,
		invitationRepo: invitationRepo,
	}
}

func (a *AdminInvitationAdapter) CheckActiveInvitationExists(ctx context.Context, email *shareddomain.Email) (bool, error) {
	_, err := a.invitationRepo.GetActiveInvitationByEmail(ctx, email.Value())
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (a *AdminInvitationAdapter) SaveInvitation(ctx context.Context, email *shareddomain.Email, inviterID string) (*adminout.SavedInvitation, error) {
	res, err := a.saveInvitation.Execute(ctx, email.Value(), inviterID)
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

package adminexternal

import (
	"context"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

var _ adminout.AdminIdentityPort = (*AdminIdentityAdapter)(nil)

type AdminIdentityAdapter struct {
	getUserByEmail identityin.GetUserByEmailUsecaseInt
}

func NewAdminIdentityAdapter(getUserByEmail identityin.GetUserByEmailUsecaseInt) *AdminIdentityAdapter {
	return &AdminIdentityAdapter{
		getUserByEmail: getUserByEmail,
	}
}

func (a *AdminIdentityAdapter) CheckUserExists(ctx context.Context, email *shareddomain.Email) (bool, error) {
	user, err := a.getUserByEmail.Execute(ctx, email.Value())
	if err != nil {
		// A missing user is the expected "does not exist" answer, not a failure.
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return user != nil, nil
}

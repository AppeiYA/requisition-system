package requisitionexternal

import (
	"context"
	"errors"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	requisitionexternal "github.com/AppeiYA/requisition-system/internal/requisition/ports/out/external"
	identitypostgres "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres"
)

type IdentityAdapter struct {
	userRepo identityout.UsersRepository
}

func NewIdentityAdapter(userRepo identityout.UsersRepository) *IdentityAdapter {
	return &IdentityAdapter{userRepo: userRepo}
}

func (a *IdentityAdapter) CheckUserExists(ctx context.Context, userID string) (bool, error) {
	user, err := a.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, identitypostgres.ErrUserNotFound) {
			return false, nil
		}
		return false, err
	}
	if user == nil {
		return false, nil
	}
	return true, nil
}

var _ requisitionexternal.UserPort = (*IdentityAdapter)(nil)

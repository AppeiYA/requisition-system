package requisitionexternal

import (
	"context"
	"errors"

	identitypostgres "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	requisitionexternal "github.com/AppeiYA/requisition-system/internal/requisition/ports/out/external"
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

func (a *IdentityAdapter) GetUser(ctx context.Context, userID string) (*requisitionout.User, error) {
	user, err := a.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, identitypostgres.ErrUserNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return &requisitionout.User{
		ID:        user.ID(),
		FirstName: user.FirstName().Value(),
		LastName:  user.LastName().Value(),
		Email:     user.Email().Value(),
		Role:      user.Role().String(),
		Status:    user.Status().Value(),
	}, nil
}

var _ requisitionexternal.UserPort = (*IdentityAdapter)(nil)
var _ requisitionout.UserLookup = (*IdentityAdapter)(nil)

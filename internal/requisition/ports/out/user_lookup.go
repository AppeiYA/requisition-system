package requisitionout

import (
	"context"
)

type User struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Role      string
	Status    string
}

type UserLookup interface {
	CheckUserExists(ctx context.Context, userID string) (bool, error)
	GetUser(ctx context.Context, userID string) (*User, error)
}

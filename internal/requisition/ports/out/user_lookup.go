package requisitionout

import (
	"context"
)

type UserLookup interface {
	CheckUserExists(ctx context.Context, userID string) (bool, error)
}

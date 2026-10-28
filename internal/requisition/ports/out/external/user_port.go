package requisitionexternal

import (
	"context"
)

type UserPort interface {
	CheckUserExists(ctx context.Context, userID string) (bool, error)
}

package identityin

import (
	"context"
)

type GetUserByEmailUsecaseInt interface {
	Execute(ctx context.Context, email string) (*UserResponse, error)
}

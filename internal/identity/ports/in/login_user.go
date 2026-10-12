package identityin

import "context"

type LoginUserRequest struct {
	Email    string
	Password string
}

type LoginUserUsecaseInt interface {
	Execute(ctx context.Context, req LoginUserRequest) (string, error)
}
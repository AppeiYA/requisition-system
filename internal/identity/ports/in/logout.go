package identityin

import "context"

type LogoutRequest struct {
	RefreshToken string
}

type LogoutUsecaseInt interface {
	Execute(ctx context.Context, req LogoutRequest) error
}

package identityin

import "context"

type RefreshTokenRequest struct {
	RefreshToken string
}

type RefreshTokenUsecaseInt interface {
	Execute(ctx context.Context, req RefreshTokenRequest) (*TokenResponse, error)
}

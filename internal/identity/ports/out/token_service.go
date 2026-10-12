package identityout

import "context"

type TokenModel struct {
	ID string
	Email string
}

type TokenService interface {
	GenerateToken(ctx context.Context, user *TokenModel) (string, error)
	ValidateToken(ctx context.Context, token string) (*TokenModel, error)
}
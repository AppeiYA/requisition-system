package identityout

import "context"

// TokenModel is the authenticated identity recovered from a validated access
// token.
type TokenModel struct {
	ID string
}

// AccessToken is a signed, short-lived JWT together with its lifetime in
// seconds, as expected by OAuth2-style responses.
type AccessToken struct {
	Value     string
	ExpiresIn int64
}

// AccessTokenService issues and validates short-lived JWT access tokens.
type AccessTokenService interface {
	GenerateAccessToken(ctx context.Context, user *TokenModel) (*AccessToken, error)
	ValidateAccessToken(ctx context.Context, token string) (*TokenModel, error)
}

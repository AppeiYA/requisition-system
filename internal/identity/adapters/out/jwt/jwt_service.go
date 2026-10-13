package jwtadapter

import (
	"context"
	"errors"
	"time"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/config"
	"github.com/golang-jwt/jwt/v5"
)

func NewJWTService(cfg config.JWTConfig) *JWTService {
	return &JWTService{
		secret:   []byte(cfg.Secret),
		ttl:      cfg.AccessTokenTTL,
		issuer:   cfg.Issuer,
		audience: cfg.Audience,
		leeway:   cfg.Leeway,
	}
}

// GenerateAccessToken signs a short-lived JWT for the given user. Only access
// tokens are JWTs; refresh tokens are opaque and persisted by the session
// repository.
func (s *JWTService) GenerateAccessToken(
	ctx context.Context,
	user *identityout.TokenModel,
) (*identityout.AccessToken, error) {
	if user == nil || user.ID == "" {
		return nil, ErrInvalidTokenModel
	}

	now := time.Now()

	claims := jwt.RegisteredClaims{
		Subject:   user.ID,
		Issuer:    s.issuer,
		Audience:  jwt.ClaimStrings{s.audience},
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return nil, signTokenError(err)
	}

	return &identityout.AccessToken{
		Value:     signedToken,
		ExpiresIn: int64(s.ttl.Seconds()),
	}, nil
}

func (s *JWTService) ValidateAccessToken(
	ctx context.Context,
	token string,
) (*identityout.TokenModel, error) {
	claims := &jwt.RegisteredClaims{}

	parsedToken, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (any, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, ErrUnexpectedSigningMethod
			}
			return s.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
		jwt.WithLeeway(s.leeway),
	)
	if err != nil {
		return nil, mapValidationError(err)
	}

	if !parsedToken.Valid || claims.Subject == "" {
		return nil, ErrInvalidToken
	}

	return &identityout.TokenModel{ID: claims.Subject}, nil
}

func mapValidationError(err error) error {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return ErrExpiredToken
	case errors.Is(err, jwt.ErrTokenMalformed):
		return ErrMalformedToken
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		return ErrInvalidTokenSignature
	default:
		return ErrInvalidToken
	}
}

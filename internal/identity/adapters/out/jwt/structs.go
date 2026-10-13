package jwtadapter

import (
	"time"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

type JWTService struct {
	secret   []byte
	ttl      time.Duration
	issuer   string
	audience string
	leeway   time.Duration
}

var _ identityout.AccessTokenService = (*JWTService)(nil)

package usecase

import (
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

// newTokenResponse maps a freshly issued access token (JWT) and opaque refresh
// token onto the transport-agnostic response DTO.
func newTokenResponse(access *identityout.AccessToken, refreshToken string) *identityin.TokenResponse {
	return &identityin.TokenResponse{
		AccessToken:  access.Value,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    access.ExpiresIn,
	}
}

package identityhttp

import (
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

// The blank assignments keep the imported types referenced so swaggo can
// resolve them from the annotations below; they have no runtime effect.
var (
	_ = response.Response{}
	_ = identityin.TokenResponse{}
)

// refreshTokenDoc documents the refresh handler in refresh.go.
//
//	@Summary		Refresh an access token
//	@Description	Exchanges a valid refresh token for a new access/refresh token pair.
//	@Tags			identity
//	@Accept			json
//	@Produce		json
//	@Param			request	body		refreshRequest	true	"Refresh token"
//	@Success		200		{object}	response.Response{data=identityin.TokenResponse}
//	@Failure		400		{object}	response.ErrorResponse	"Validation error"
//	@Failure		401		{object}	response.Response		"Invalid or expired refresh token"
//	@Router			/auth/refresh [post]
func refreshTokenDoc() {}

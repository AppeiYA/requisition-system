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

// loginUserDoc documents the login handler in login.go.
//
//	@Summary		Authenticate a user
//	@Description	Validates the given credentials and returns a new access/refresh token pair.
//	@Tags			identity
//	@Accept			json
//	@Produce		json
//	@Param			request	body		loginRequest	true	"Login credentials"
//	@Success		200		{object}	response.Response{data=identityin.TokenResponse}
//	@Failure		400		{object}	response.ErrorResponse	"Validation error"
//	@Failure		401		{object}	response.Response		"Invalid credentials"
//	@Router			/auth/login [post]
func loginUserDoc() {}

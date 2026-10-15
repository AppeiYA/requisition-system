package identityhttp

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

// The blank assignment keeps the imported type referenced so swaggo can resolve
// it from the annotations below; it has no runtime effect.
var _ = response.Response{}

// logoutDoc documents the logout handler in logout.go.
//
//	@Summary		Log out a user
//	@Description	Revokes the session associated with the given refresh token.
//	@Tags			identity
//	@Accept			json
//	@Produce		json
//	@Param			request	body		logoutRequest	true	"Refresh token to revoke"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.ErrorResponse	"Validation error"
//	@Failure		401		{object}	response.Response		"Invalid or expired refresh token"
//	@Router			/auth/logout [post]
func logoutDoc() {}

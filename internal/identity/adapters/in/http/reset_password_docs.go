package identityhttp

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

var _ = response.Response{}

// resetPasswordDoc documents the reset-password handler.
//
//	@Summary		Reset a password with a reset token
//	@Description	Sets a new password using a single-use token from the reset email. Revokes all refresh sessions and does not issue new credentials.
//	@Tags			identity
//	@Accept			json
//	@Produce		json
//	@Param			request	body		resetPasswordRequest	true	"Reset token and new password"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.ErrorResponse	"Validation error, weak password, or invalid/expired token"
//	@Failure		429		{object}	response.Response		"Too many requests"
//	@Router			/auth/reset-password [post]
func resetPasswordDoc() {}

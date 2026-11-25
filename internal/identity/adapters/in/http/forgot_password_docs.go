package identityhttp

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

var _ = response.Response{}

// forgotPasswordDoc documents the forgot-password handler.
//
//	@Summary		Request a password reset link
//	@Description	Always returns the same response, whether or not the account exists, is active, or is throttled. No account information is revealed.
//	@Tags			identity
//	@Accept			json
//	@Produce		json
//	@Param			request	body		forgotPasswordRequest	true	"Email address"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.ErrorResponse	"Validation error"
//	@Failure		429		{object}	response.Response		"Too many requests"
//	@Router			/auth/forgot-password [post]
func forgotPasswordDoc() {}

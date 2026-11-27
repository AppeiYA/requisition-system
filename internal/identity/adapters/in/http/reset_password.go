package identityhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

func (h *Handler) resetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithDetails(
			c,
			"VALIDATION_ERROR",
			"validation failed",
			http.StatusBadRequest,
			response.ValidationErrors(err),
		)
		return
	}

	if err := h.identityPortIn.ResetPasswordUsecase.Execute(c.Request.Context(), req.Token, req.Password); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// No token, user, or new credentials are returned. The client sends the
	// user to the login page.
	response.Success(c, "password reset successful", http.StatusOK, nil)
}

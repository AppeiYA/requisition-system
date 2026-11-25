package identityhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

const forgotPasswordMessage = "if an account with that email exists, a password reset link has been sent"

func (h *Handler) forgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
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

	// The use case returns nil for unknown, inactive and throttled accounts,
	// so this response is identical in every one of those cases.
	if err := h.identityPortIn.RequestPasswordResetUsecase.Execute(c.Request.Context(), req.Email); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, forgotPasswordMessage, http.StatusOK, nil)
}

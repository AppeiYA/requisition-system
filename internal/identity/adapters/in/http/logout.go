package identityhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/response"
)

func (h *Handler) logout(c *gin.Context) {
	var req logoutRequest
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

	if err := h.logoutUsecase.Execute(c.Request.Context(), identityin.LogoutRequest{
		RefreshToken: req.RefreshToken,
	}); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, "logout successful", http.StatusOK, nil)
}

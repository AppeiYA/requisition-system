package identityhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) refresh(c *gin.Context) {
	var req refreshRequest
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

	tokens, err := h.refreshUsecase.Execute(c.Request.Context(), identityin.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, "token refreshed", http.StatusOK, tokens)
}

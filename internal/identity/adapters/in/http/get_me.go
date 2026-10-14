package identityhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

func (h *Handler) getMe(c *gin.Context) {
	userID := c.GetString(userIDContextKey)

	user, err := h.getMeUsecase.Execute(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, "user retrieved", http.StatusOK, ToJSON(user))
}

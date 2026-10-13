package identityhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/response"
)

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
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

	tokens, err := h.loginUsecase.Execute(c.Request.Context(), identityin.LoginUserRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, "login successful", http.StatusOK, tokens)
}

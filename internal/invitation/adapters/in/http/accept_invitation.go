package invitationhttp

import (
	"net/http"

	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *InvitationHandler) AcceptInvitation(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		response.ErrorWithDetails(c, "validation_error", "token is required", http.StatusBadRequest, nil)
		c.Abort()
		return
	}
	var req AcceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, response.ValidationErrors(err))
		c.Abort()
		return
	}

	err := h.acceptInvitationUC.Execute(c.Request.Context(), invitationin.AcceptInvitationCommand{
		Token:     token,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  req.Password,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "registration completed", http.StatusOK, nil)
}

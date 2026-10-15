package invitationhttp

import (
	"net/http"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *InvitationHandler) ValidateInvitation(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		response.ErrorWithDetails(c, "validation_error", "token is required", http.StatusBadRequest, nil)
		c.Abort()
		return
	}

	result, err := h.validateInvitationUC.Execute(c.Request.Context(), token)
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "invitation validated", http.StatusOK, ValidateInvitationResponse{
		Valid:        result.Valid,
		Email:        result.Email,
		FirstName:    result.FirstName,
		LastName:     result.LastName,
		Phone:        result.Phone,
		StaffID:      result.StaffID,
		DepartmentID: result.DepartmentID,
		JobTitle:     result.JobTitle,
	})
}

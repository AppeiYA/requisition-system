package adminhttp

import (
	"net/http"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/authcontext"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) invite(c *gin.Context) {
	userID, ok := authcontext.GetUserID(c)
	if !ok {
		response.Error(c, "unauthorized", http.StatusUnauthorized, nil)
		c.Abort()
		return
	}

	var req InviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, response.ValidationErrors(err))
		c.Abort()
		return
	}

	res, err := h.adminPortIn.InviteMember.Execute(c.Request.Context(), adminin.InviteMemberCommand{
		InviterID:    userID,
		Email:        req.Email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		StaffID:      req.StaffID,
		DepartmentID: req.DepartmentID,
		JobTitle:     req.JobTitle,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "invitation sent", http.StatusCreated, res)
}
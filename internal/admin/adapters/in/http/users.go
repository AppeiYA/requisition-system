package adminhttp

import (
	"net/http"
	"strings"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/authcontext"
	apppagination "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/pagination"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) listUsers(c *gin.Context) {
	page, ok := apppagination.Parse(c)
	if !ok {
		return
	}

	filter := adminin.UserListFilter{
		Search:       strings.TrimSpace(c.Query("q")),
		Role:         c.Query("role"),
		Status:       c.Query("status"),
		DepartmentID: c.Query("department_id"),
		Sort:         c.Query("sort"),
		Limit:        page.Limit,
		Offset:       page.Offset,
	}

	var details []response.FieldError
	if len(filter.Search) > 100 {
		details = append(details, response.FieldError{Field: "q", Message: "must be at most 100 characters"})
	}
	if filter.Role != "" && !shareddomain.IsValidRole(filter.Role) {
		details = append(details, response.FieldError{Field: "role", Message: "must be admin or user"})
	}
	if filter.Status != "" {
		if !adminin.ValidUserStatuses[filter.Status] {
			details = append(details, response.FieldError{Field: "status", Message: "must be active, suspended or inactive"})
		}
	}
	if filter.DepartmentID != "" && !strings.HasPrefix(filter.DepartmentID, departmentIDPrefix+"-") {
		details = append(details, response.FieldError{Field: "department_id", Message: "must be a valid department ID"})
	}
	if !adminin.ValidUserSortValues[filter.Sort] {
		details = append(details, response.FieldError{Field: "sort", Message: "unsupported sort value"})
	}
	if len(details) > 0 {
		response.ErrorWithDetails(c, "validation_error", "invalid user filters", http.StatusBadRequest, details)
		return
	}

	result, err := h.adminPortIn.ListUsers.Execute(c.Request.Context(), adminin.ListUsersQuery{Filter: filter})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.SuccessWithPaginationTotal(
		c,
		"users fetched successfully",
		http.StatusOK,
		toUserSummaryResponses(result.Users),
		response.NewPaginationMetaWithTotal(page.Limit, page.Offset, result.Total),
	)
}

func (h *Handler) getUser(c *gin.Context) {
	userID := c.Param("id")
	if !strings.HasPrefix(userID, "usr-") {
		response.ErrorWithDetails(c, "validation_error", "invalid user id", http.StatusBadRequest, []response.FieldError{
			{Field: "id", Message: "must be a valid user ID"},
		})
		return
	}

	user, err := h.adminPortIn.GetUser.Execute(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, "user fetched successfully", http.StatusOK, toUserDetailResponse(*user))
}

func (h *Handler) updateUserStatus(c *gin.Context) {
	actorID, ok := authcontext.GetUserID(c)
	if !ok {
		response.Error(c, "unauthorized", http.StatusUnauthorized, nil)
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, response.ValidationErrors(err))
		return
	}

	err := h.adminPortIn.UpdateUserStatus.Execute(c.Request.Context(), adminin.UpdateUserStatusCommand{
		UserID:  c.Param("id"),
		Status:  req.Status,
		Reason:  req.Reason,
		ActorID: actorID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, "user status updated", http.StatusOK, nil)
}

func (h *Handler) updateUserRole(c *gin.Context) {
	actorID, ok := authcontext.GetUserID(c)
	if !ok {
		response.Error(c, "unauthorized", http.StatusUnauthorized, nil)
		return
	}

	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, response.ValidationErrors(err))
		return
	}

	err := h.adminPortIn.UpdateUserRole.Execute(c.Request.Context(), adminin.UpdateUserRoleCommand{
		UserID:  c.Param("id"),
		Role:    req.Role,
		Reason:  req.Reason,
		ActorID: actorID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, "user role updated", http.StatusOK, nil)
}

func (h *Handler) revokeUserSessions(c *gin.Context) {
	actorID, ok := authcontext.GetUserID(c)
	if !ok {
		response.Error(c, "unauthorized", http.StatusUnauthorized, nil)
		return
	}

	err := h.adminPortIn.RevokeUserSessions.Execute(c.Request.Context(), adminin.RevokeUserSessionsCommand{
		UserID:  c.Param("id"),
		Reason:  strings.TrimSpace(c.Query("reason")),
		ActorID: actorID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, "user sessions revoked", http.StatusOK, nil)
}

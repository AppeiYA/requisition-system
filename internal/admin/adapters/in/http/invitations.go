package adminhttp

import (
	"net/http"
	"strings"
	"time"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/authcontext"
	apppagination "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/pagination"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

// The HTTP layer validates ID shapes itself rather than importing another
// module's domain constants for a transport-level check.
const (
	departmentIDPrefix = "dpt"
	invitationIDPrefix = "inv"
	userIDPrefix       = "usr"
)

func parseInvitationDate(field, raw string) (*time.Time, []response.FieldError) {
	if raw == "" {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			utc := parsed.UTC()
			return &utc, nil
		}
	}
	return nil, []response.FieldError{{Field: field, Message: "must be an RFC3339 date or date-time"}}
}

func (h *Handler) listInvitations(c *gin.Context) {
	page, ok := apppagination.Parse(c)
	if !ok {
		return
	}

	filter := adminin.InvitationListFilter{
		Search:       strings.TrimSpace(c.Query("q")),
		Status:       c.Query("status"),
		DepartmentID: c.Query("department_id"),
		InvitedBy:    c.Query("invited_by"),
		Limit:        page.Limit,
		Offset:       page.Offset,
	}

	var details []response.FieldError
	if len(filter.Search) > 100 {
		details = append(details, response.FieldError{Field: "q", Message: "must be at most 100 characters"})
	}
	if filter.Status != "" && !adminin.ValidInvitationStatuses[filter.Status] {
		details = append(details, response.FieldError{Field: "status", Message: "must be a valid invitation status"})
	}
	if filter.DepartmentID != "" && !strings.HasPrefix(filter.DepartmentID, departmentIDPrefix+"-") {
		details = append(details, response.FieldError{Field: "department_id", Message: "must be a valid department ID"})
	}
	if filter.InvitedBy != "" && !strings.HasPrefix(filter.InvitedBy, userIDPrefix+"-") {
		details = append(details, response.FieldError{Field: "invited_by", Message: "must be a valid user ID"})
	}

	var dateErrors []response.FieldError
	filter.From, dateErrors = parseInvitationDate("from", c.Query("from"))
	details = append(details, dateErrors...)
	filter.To, dateErrors = parseInvitationDate("to", c.Query("to"))
	details = append(details, dateErrors...)
	if filter.From != nil && filter.To != nil && filter.From.After(*filter.To) {
		details = append(details, response.FieldError{Field: "from", Message: "must not be after to"})
	}
	if len(details) > 0 {
		response.ErrorWithDetails(c, "validation_error", "invalid invitation filters", http.StatusBadRequest, details)
		return
	}

	result, err := h.adminPortIn.ListInvitations.Execute(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.SuccessWithPaginationTotal(
		c,
		"invitations fetched successfully",
		http.StatusOK,
		toInvitationSummaryResponses(result.Invitations),
		response.NewPaginationMetaWithTotal(page.Limit, page.Offset, result.Total),
	)
}

func (h *Handler) getInvitation(c *gin.Context) {
	invitationID := c.Param("id")
	if !strings.HasPrefix(invitationID, invitationIDPrefix+"-") {
		response.ErrorWithDetails(c, "validation_error", "invalid invitation id", http.StatusBadRequest, []response.FieldError{
			{Field: "id", Message: "must be a valid invitation ID"},
		})
		return
	}

	invitation, err := h.adminPortIn.GetInvitation.Execute(c.Request.Context(), invitationID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, "invitation fetched successfully", http.StatusOK, toInvitationDetailResponse(*invitation))
}

func (h *Handler) resendInvitation(c *gin.Context) {
	actorID, ok := authcontext.GetUserID(c)
	if !ok {
		response.Error(c, "unauthorized", http.StatusUnauthorized, nil)
		return
	}

	invitationID := c.Param("id")
	if !strings.HasPrefix(invitationID, invitationIDPrefix+"-") {
		response.ErrorWithDetails(c, "validation_error", "invalid invitation id", http.StatusBadRequest, []response.FieldError{
			{Field: "id", Message: "must be a valid invitation ID"},
		})
		return
	}

	result, err := h.adminPortIn.ResendInvitation.Execute(c.Request.Context(), adminin.ResendInvitationCommand{
		InvitationID: invitationID,
		ActorID:      actorID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, "invitation resent", http.StatusOK, ResendInvitationResponse{ID: result.ID, ExpiresAt: result.ExpiresAt})
}

func (h *Handler) revokeInvitation(c *gin.Context) {
	actorID, ok := authcontext.GetUserID(c)
	if !ok {
		response.Error(c, "unauthorized", http.StatusUnauthorized, nil)
		return
	}

	invitationID := c.Param("id")
	if !strings.HasPrefix(invitationID, invitationIDPrefix+"-") {
		response.ErrorWithDetails(c, "validation_error", "invalid invitation id", http.StatusBadRequest, []response.FieldError{
			{Field: "id", Message: "must be a valid invitation ID"},
		})
		return
	}

	reason := strings.TrimSpace(c.Query("reason"))
	if len([]rune(reason)) > 500 {
		response.ErrorWithDetails(c, "validation_error", "invalid revoke reason", http.StatusBadRequest, []response.FieldError{
			{Field: "reason", Message: "must be at most 500 characters"},
		})
		return
	}

	err := h.adminPortIn.RevokeInvitation.Execute(c.Request.Context(), adminin.RevokeInvitationCommand{
		InvitationID: invitationID,
		ActorID:      actorID,
		Reason:       reason,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, "invitation revoked", http.StatusOK, nil)
}

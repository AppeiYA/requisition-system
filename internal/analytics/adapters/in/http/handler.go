package analyticshttp

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	analyticdomain "github.com/AppeiYA/requisition-system/internal/analytics/domain"
	analyticsin "github.com/AppeiYA/requisition-system/internal/analytics/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	portIn *analyticsin.PortIn
}

func NewHandler(portIn *analyticsin.PortIn) *Handler {
	return &Handler{portIn: portIn}
}

func (h *Handler) Register(router *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	admin := router.Group("/admin/analytics")
	admin.Use(middlewares...)
	admin.GET("/overview", h.overview)
	admin.GET("/requisitions", h.requisitions)
	admin.GET("/users", h.users)
	admin.GET("/catalogue", h.catalogue)
}

func parseFilter(c *gin.Context) (analyticdomain.Filter, []response.FieldError) {
	now := time.Now().UTC()
	to := now
	from := now.AddDate(0, 0, -30)
	var details []response.FieldError

	if raw := c.Query("from"); raw != "" {
		parsed, ok := parseAnalyticsDate(raw)
		if !ok {
			details = append(details, response.FieldError{Field: "from", Message: "must be an RFC3339 date or date-time"})
		} else {
			from = parsed
		}
	}
	if raw := c.Query("to"); raw != "" {
		parsed, ok := parseAnalyticsDate(raw)
		if !ok {
			details = append(details, response.FieldError{Field: "to", Message: "must be an RFC3339 date or date-time"})
		} else {
			to = parsed
		}
	}
	if len(details) == 0 {
		if from.After(to) {
			details = append(details, response.FieldError{Field: "from", Message: "must be before to"})
		}
		if to.Sub(from) > 366*24*time.Hour {
			details = append(details, response.FieldError{Field: "to", Message: "range must not exceed 366 days"})
		}
	}

	interval := c.Query("interval")
	if interval == "" {
		interval = "day"
	}
	if interval != "day" && interval != "week" && interval != "month" {
		details = append(details, response.FieldError{Field: "interval", Message: "must be day, week or month"})
	}

	limit := 10
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 50 {
			details = append(details, response.FieldError{Field: "limit", Message: "must be between 1 and 50"})
		} else {
			limit = parsed
		}
	}

	departmentID := strings.TrimSpace(c.Query("department_id"))
	if departmentID != "" && !strings.HasPrefix(departmentID, "dpt-") {
		details = append(details, response.FieldError{Field: "department_id", Message: "must be a valid department ID"})
	}

	return analyticdomain.Filter{
		From: from, To: to, DepartmentID: departmentID, Interval: interval, Limit: limit,
	}, details
}

func parseAnalyticsDate(raw string) (time.Time, bool) {
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}

func (h *Handler) overview(c *gin.Context) {
	filter, details := parseFilter(c)
	if len(details) > 0 {
		response.ErrorWithDetails(c, "validation_error", "invalid analytics filters", http.StatusBadRequest, details)
		return
	}
	result, err := h.portIn.Overview.ExecuteOverview(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, "analytics overview fetched successfully", http.StatusOK, toOverviewResponse(result))
}

func (h *Handler) requisitions(c *gin.Context) {
	h.executeRequisitions(c)
}

func (h *Handler) executeRequisitions(c *gin.Context) {
	filter, details := parseFilter(c)
	if len(details) > 0 {
		response.ErrorWithDetails(c, "validation_error", "invalid analytics filters", http.StatusBadRequest, details)
		return
	}
	result, err := h.portIn.Requisitions.ExecuteRequisitions(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, "requisition analytics fetched successfully", http.StatusOK, toRequisitionAnalyticsResponse(result))
}

func (h *Handler) users(c *gin.Context) {
	filter, details := parseFilter(c)
	if len(details) > 0 {
		response.ErrorWithDetails(c, "validation_error", "invalid analytics filters", http.StatusBadRequest, details)
		return
	}
	result, err := h.portIn.Users.ExecuteUsers(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, "user analytics fetched successfully", http.StatusOK, toUserAnalyticsResponse(result))
}

func (h *Handler) catalogue(c *gin.Context) {
	filter, details := parseFilter(c)
	if len(details) > 0 {
		response.ErrorWithDetails(c, "validation_error", "invalid analytics filters", http.StatusBadRequest, details)
		return
	}
	result, err := h.portIn.Catalogue.ExecuteCatalogue(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, "catalogue analytics fetched successfully", http.StatusOK, toCatalogueAnalyticsResponse(result))
}

func toOverviewResponse(value *analyticdomain.Overview) OverviewResponse {
	result := OverviewResponse{GeneratedAt: value.GeneratedAt, Range: RangeResponse{From: value.Range.From, To: value.Range.To}}
	result.DataAsOf = value.DataAsOf
	result.Source = value.Source
	result.Users.Total = value.Users.Total
	result.Users.Active = value.Users.Active
	result.Users.Suspended = value.Users.Suspended
	result.Users.Inactive = value.Users.Inactive
	result.Users.Admins = value.Users.Admins
	result.Users.NewInRange = value.Users.NewInRange
	result.Invitations.Total = value.Invitations.Total
	result.Invitations.Pending = value.Invitations.Pending
	result.Invitations.Delivered = value.Invitations.Delivered
	result.Invitations.Accepted = value.Invitations.Accepted
	result.Invitations.Expired = value.Invitations.Expired
	result.Invitations.Revoked = value.Invitations.Revoked
	result.Invitations.Failed = value.Invitations.Failed
	result.Invitations.AcceptanceRate = value.Invitations.AcceptanceRate
	result.Requisitions.Total = value.Requisitions.Total
	result.Requisitions.ByStatus = value.Requisitions.ByStatus
	result.Requisitions.TotalEstimatedAmount = value.Requisitions.TotalEstimatedAmount
	result.Requisitions.AverageEstimatedAmount = value.Requisitions.AverageEstimatedAmount
	result.Requisitions.Currency = value.Requisitions.Currency
	result.Requisitions.PendingOverdue = value.Requisitions.PendingOverdue
	result.Catalogue.ActiveItems = value.Catalogue.ActiveItems
	result.Catalogue.ActiveCategories = value.Catalogue.ActiveCategories
	result.Catalogue.UnreviewedRequestedItems = value.Catalogue.UnreviewedRequestedItems
	return result
}

func toRequisitionAnalyticsResponse(value *analyticdomain.RequisitionAnalytics) RequisitionAnalyticsResponse {
	result := RequisitionAnalyticsResponse{
		GeneratedAt: value.GeneratedAt, Range: RangeResponse{From: value.Range.From, To: value.Range.To},
		ByStatus: breakdownResponses(value.ByStatus), ByPriority: breakdownResponses(value.ByPriority),
		ByDepartment: breakdownResponses(value.ByDepartment), ByCategory: breakdownResponses(value.ByCategory),
		Trend: trendResponses(value.Trend), TopItems: breakdownResponses(value.TopItems),
		TopRequesters: breakdownResponses(value.TopRequesters),
	}
	result.DataAsOf = value.DataAsOf
	result.Source = value.Source
	result.CycleTime.PendingToApprovedSeconds = value.CycleTime.PendingToApprovedSeconds
	result.CycleTime.ApprovedToCompletedSeconds = value.CycleTime.ApprovedToCompletedSeconds
	result.CycleTime.CreationToClosedSeconds = value.CycleTime.CreationToClosedSeconds
	return result
}

func toUserAnalyticsResponse(value *analyticdomain.UserAnalytics) UserAnalyticsResponse {
	result := UserAnalyticsResponse{
		GeneratedAt: value.GeneratedAt, Range: RangeResponse{From: value.Range.From, To: value.Range.To},
		Growth: trendResponses(value.Growth), TopRequesters: breakdownResponses(value.TopRequesters),
	}
	result.DataAsOf = value.DataAsOf
	result.Source = value.Source
	result.ByRole = make([]CountValueResponse, len(value.ByRole))
	for i, item := range value.ByRole {
		result.ByRole[i] = CountValueResponse{Name: item.Name, Count: item.Count}
	}
	result.ByStatus = make([]CountValueResponse, len(value.ByStatus))
	for i, item := range value.ByStatus {
		result.ByStatus[i] = CountValueResponse{Name: item.Name, Count: item.Count}
	}
	result.ByDepartment = make([]struct {
		Name          string `json:"name"`
		Members       int    `json:"members"`
		ActiveMembers int    `json:"active_members"`
	}, len(value.ByDepartment))
	for i, item := range value.ByDepartment {
		result.ByDepartment[i].Name = item.Name
		result.ByDepartment[i].Members = item.Members
		result.ByDepartment[i].ActiveMembers = item.ActiveMembers
	}
	result.InvitationFunnel.Created = value.InvitationFunnel.Created
	result.InvitationFunnel.Delivered = value.InvitationFunnel.Delivered
	result.InvitationFunnel.Accepted = value.InvitationFunnel.Accepted
	result.InvitationFunnel.Expired = value.InvitationFunnel.Expired
	result.InvitationFunnel.Revoked = value.InvitationFunnel.Revoked
	result.InvitationFunnel.Failed = value.InvitationFunnel.Failed
	result.InvitationFunnel.AcceptanceRate = value.InvitationFunnel.AcceptanceRate
	result.InvitationFunnel.AverageTimeToAcceptSeconds = value.InvitationFunnel.AverageTimeToAcceptSeconds
	return result
}

func toCatalogueAnalyticsResponse(value *analyticdomain.CatalogueAnalytics) CatalogueAnalyticsResponse {
	result := CatalogueAnalyticsResponse{
		GeneratedAt: value.GeneratedAt, Range: RangeResponse{From: value.Range.From, To: value.Range.To},
		ActiveItems: value.ActiveItems, ActiveCategories: value.ActiveCategories,
		InactiveItems: value.InactiveItems, InactiveCategories: value.InactiveCategories,
		TopItems: breakdownResponses(value.TopItems), TopCategories: breakdownResponses(value.TopCategories),
	}
	result.DataAsOf = value.DataAsOf
	result.Source = value.Source
	result.RequestedItemReview.Unreviewed = value.RequestedItemReview.Unreviewed
	result.RequestedItemReview.Promoted = value.RequestedItemReview.Promoted
	result.RequestedItemReview.Dismissed = value.RequestedItemReview.Dismissed
	result.ReviewBacklogAging.OlderThan7Days = value.ReviewBacklogAging.OlderThan7Days
	result.ReviewBacklogAging.OlderThan30Days = value.ReviewBacklogAging.OlderThan30Days
	return result
}

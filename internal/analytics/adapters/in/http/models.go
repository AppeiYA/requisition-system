package analyticshttp

import (
	"time"

	analyticdomain "github.com/AppeiYA/requisition-system/internal/analytics/domain"
)

type RangeResponse struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}
type CountValueResponse struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
type BreakdownResponse struct {
	Name            string  `json:"name"`
	Count           int     `json:"count"`
	EstimatedAmount float64 `json:"estimated_amount"`
}
type TrendResponse struct {
	Bucket          time.Time `json:"bucket"`
	Count           int       `json:"count"`
	EstimatedAmount float64   `json:"estimated_amount,omitempty"`
}

type OverviewResponse struct {
	GeneratedAt time.Time     `json:"generated_at"`
	DataAsOf    time.Time     `json:"data_as_of"`
	Source      string        `json:"source"`
	Range       RangeResponse `json:"range"`
	Users       struct {
		Total      int `json:"total"`
		Active     int `json:"active"`
		Suspended  int `json:"suspended"`
		Inactive   int `json:"inactive"`
		Admins     int `json:"admins"`
		NewInRange int `json:"new_in_range"`
	} `json:"users"`
	Invitations struct {
		Total          int     `json:"total"`
		Pending        int     `json:"pending"`
		Delivered      int     `json:"delivered"`
		Accepted       int     `json:"accepted"`
		Expired        int     `json:"expired"`
		Revoked        int     `json:"revoked"`
		Failed         int     `json:"failed"`
		AcceptanceRate float64 `json:"acceptance_rate"`
	} `json:"invitations"`
	Requisitions struct {
		Total                  int            `json:"total"`
		ByStatus               map[string]int `json:"by_status"`
		TotalEstimatedAmount   float64        `json:"total_estimated_amount"`
		AverageEstimatedAmount float64        `json:"average_estimated_amount"`
		Currency               string         `json:"currency"`
		PendingOverdue         int            `json:"pending_overdue"`
	} `json:"requisitions"`
	Catalogue struct {
		ActiveItems              int `json:"active_items"`
		ActiveCategories         int `json:"active_categories"`
		UnreviewedRequestedItems int `json:"unreviewed_requested_items"`
	} `json:"catalogue"`
}

type RequisitionAnalyticsResponse struct {
	GeneratedAt  time.Time           `json:"generated_at"`
	DataAsOf     time.Time           `json:"data_as_of"`
	Source       string              `json:"source"`
	Range        RangeResponse       `json:"range"`
	ByStatus     []BreakdownResponse `json:"by_status"`
	ByPriority   []BreakdownResponse `json:"by_priority"`
	ByDepartment []BreakdownResponse `json:"by_department"`
	ByCategory   []BreakdownResponse `json:"by_category"`
	Trend        []TrendResponse     `json:"trend"`
	CycleTime    struct {
		PendingToApprovedSeconds   float64 `json:"pending_to_approved_seconds"`
		ApprovedToCompletedSeconds float64 `json:"approved_to_completed_seconds"`
		CreationToClosedSeconds    float64 `json:"creation_to_closed_seconds"`
	} `json:"cycle_time"`
	TopItems      []BreakdownResponse `json:"top_items"`
	TopRequesters []BreakdownResponse `json:"top_requesters"`
}

type UserAnalyticsResponse struct {
	GeneratedAt  time.Time            `json:"generated_at"`
	DataAsOf     time.Time            `json:"data_as_of"`
	Source       string               `json:"source"`
	Range        RangeResponse        `json:"range"`
	Growth       []TrendResponse      `json:"growth"`
	ByRole       []CountValueResponse `json:"by_role"`
	ByStatus     []CountValueResponse `json:"by_status"`
	ByDepartment []struct {
		Name          string `json:"name"`
		Members       int    `json:"members"`
		ActiveMembers int    `json:"active_members"`
	} `json:"by_department"`
	InvitationFunnel struct {
		Created                    int     `json:"created"`
		Delivered                  int     `json:"delivered"`
		Accepted                   int     `json:"accepted"`
		Expired                    int     `json:"expired"`
		Revoked                    int     `json:"revoked"`
		Failed                     int     `json:"failed"`
		AcceptanceRate             float64 `json:"acceptance_rate"`
		AverageTimeToAcceptSeconds float64 `json:"average_time_to_accept_seconds"`
	} `json:"invitation_funnel"`
	TopRequesters []BreakdownResponse `json:"top_requesters"`
}

type CatalogueAnalyticsResponse struct {
	GeneratedAt         time.Time     `json:"generated_at"`
	DataAsOf            time.Time     `json:"data_as_of"`
	Source              string        `json:"source"`
	Range               RangeResponse `json:"range"`
	ActiveItems         int           `json:"active_items"`
	ActiveCategories    int           `json:"active_categories"`
	InactiveItems       int           `json:"inactive_items"`
	InactiveCategories  int           `json:"inactive_categories"`
	RequestedItemReview struct {
		Unreviewed int `json:"unreviewed"`
		Promoted   int `json:"promoted"`
		Dismissed  int `json:"dismissed"`
	} `json:"requested_item_review"`
	TopItems           []BreakdownResponse `json:"top_items"`
	TopCategories      []BreakdownResponse `json:"top_categories"`
	ReviewBacklogAging struct {
		OlderThan7Days  int `json:"older_than_7_days"`
		OlderThan30Days int `json:"older_than_30_days"`
	} `json:"review_backlog_aging"`
}

func breakdownResponses(items []analyticdomain.Breakdown) []BreakdownResponse {
	result := make([]BreakdownResponse, len(items))
	for i, item := range items {
		result[i] = BreakdownResponse{Name: item.Name, Count: item.Count, EstimatedAmount: item.EstimatedAmount}
	}
	return result
}

func trendResponses(items []analyticdomain.TrendBucket) []TrendResponse {
	result := make([]TrendResponse, len(items))
	for i, item := range items {
		result[i] = TrendResponse{Bucket: item.Bucket, Count: item.Count, EstimatedAmount: item.EstimatedAmount}
	}
	return result
}

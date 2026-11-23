package domain

import "time"

type Filter struct {
	From         time.Time
	To           time.Time
	DepartmentID string
	Interval     string
	Limit        int
}

type Range struct {
	From time.Time
	To   time.Time
}
type CountValue struct {
	Name  string
	Count int
}
type Breakdown struct {
	Name            string
	Count           int
	EstimatedAmount float64
}
type TrendBucket struct {
	Bucket          time.Time
	Count           int
	EstimatedAmount float64
}

type Overview struct {
	GeneratedAt  time.Time
	DataAsOf     time.Time
	Source       string
	Range        Range
	Users        OverviewUsers
	Invitations  OverviewInvitations
	Requisitions OverviewRequisitions
	Catalogue    OverviewCatalogue
}

type OverviewUsers struct{ Total, Active, Suspended, Inactive, Admins, NewInRange int }
type OverviewInvitations struct {
	Total, Pending, Delivered, Accepted, Expired, Revoked, Failed int
	AcceptanceRate                                                float64
}
type OverviewRequisitions struct {
	Total                  int
	ByStatus               map[string]int
	TotalEstimatedAmount   float64
	AverageEstimatedAmount float64
	Currency               string
	PendingOverdue         int
}
type OverviewCatalogue struct{ ActiveItems, ActiveCategories, UnreviewedRequestedItems int }

type CycleTime struct{ PendingToApprovedSeconds, ApprovedToCompletedSeconds, CreationToClosedSeconds float64 }

type RequisitionAnalytics struct {
	GeneratedAt   time.Time
	DataAsOf      time.Time
	Source        string
	Range         Range
	ByStatus      []Breakdown
	ByPriority    []Breakdown
	ByDepartment  []Breakdown
	ByCategory    []Breakdown
	Trend         []TrendBucket
	CycleTime     CycleTime
	TopItems      []Breakdown
	TopRequesters []Breakdown
}

type InvitationFunnel struct {
	Created, Delivered, Accepted, Expired, Revoked, Failed int
	AcceptanceRate                                         float64
	AverageTimeToAcceptSeconds                             float64
}

type DepartmentBreakdown struct {
	Name                   string
	Members, ActiveMembers int
}

type UserAnalytics struct {
	GeneratedAt      time.Time
	DataAsOf         time.Time
	Source           string
	Range            Range
	Growth           []TrendBucket
	ByRole           []CountValue
	ByStatus         []CountValue
	ByDepartment     []DepartmentBreakdown
	InvitationFunnel InvitationFunnel
	TopRequesters    []Breakdown
}

type ReviewBacklogAging struct{ OlderThan7Days, OlderThan30Days int }
type ReviewStatusCounts struct{ Unreviewed, Promoted, Dismissed int }

type CatalogueAnalytics struct {
	GeneratedAt                                                      time.Time
	DataAsOf                                                         time.Time
	Source                                                           string
	Range                                                            Range
	ActiveItems, ActiveCategories, InactiveItems, InactiveCategories int
	RequestedItemReview                                              ReviewStatusCounts
	TopItems                                                         []Breakdown
	TopCategories                                                    []Breakdown
	ReviewBacklogAging                                               ReviewBacklogAging
}

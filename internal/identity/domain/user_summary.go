package domain

import "time"

// UserListFilter is the validated set of filters for the admin user list.
// Validation of free-text length and enum values happens at the HTTP boundary;
// the repository relies on Sort being one of the values below.
type UserListFilter struct {
	Search       string
	Role         string
	Status       string
	DepartmentID string
	Sort         string
	Limit        int
	Offset       int
}

// ValidUserSortValues is the allow-list for the sort query parameter. An
// allow-list is used instead of interpolating the parameter so the SQL cannot
// be influenced by the caller.
var ValidUserSortValues = map[string]bool{
	"":            true,
	"created_at":  true,
	"-created_at": true,
	"last_name":   true,
	"email":       true,
}

// UserSummary is the read model for the admin user list. It deliberately has
// no password hash: that value must never leave the identity module.
type UserSummary struct {
	ID               string
	FirstName        string
	LastName         string
	Email            string
	StaffID          string
	Phone            string
	Role             string
	Status           string
	DepartmentID     string
	DepartmentName   string
	JobTitle         string
	MembershipStatus string
	CreatedAt        time.Time
	UpdatedAt        *time.Time
}

// UserDetail adds requisition activity to a user summary. Money is in major
// NGN, matching the rest of the API; the database stores kobo.
type UserDetail struct {
	UserSummary
	RequisitionCounts    map[string]int
	TotalEstimatedAmount float64
	LastRequisitionAt    *time.Time
}

// RequisitionStatuses are the keys always present in RequisitionCounts, so
// consumers do not have to special-case an empty bucket.
var RequisitionStatuses = []string{
	"pending", "approved", "in_fulfillment", "completed", "rejected", "cancelled",
}

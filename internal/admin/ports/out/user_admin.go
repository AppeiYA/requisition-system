package adminout

import (
	"context"
	"time"
)

// UserAdminPort is the admin module's view of user administration.

type UserListFilter struct {
	Search       string
	Role         string
	Status       string
	DepartmentID string
	Sort         string
	Limit        int
	Offset       int
}

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

type UserDetail struct {
	UserSummary
	RequisitionCounts    map[string]int
	TotalEstimatedAmount float64
	LastRequisitionAt    *time.Time
}

type UserAdminPort interface {
	ListUsers(ctx context.Context, filter UserListFilter) ([]UserSummary, int, error)
	GetUser(ctx context.Context, userID string) (*UserDetail, error)
	UpdateUserStatus(ctx context.Context, userID, status, reason, actorID string) error
	UpdateUserRole(ctx context.Context, userID, role, reason, actorID string) error
	RevokeUserSessions(ctx context.Context, userID, reason, actorID string) error
}

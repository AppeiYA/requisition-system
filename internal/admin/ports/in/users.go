package adminin

import (
	"context"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
)

var ValidUserSortValues = map[string]bool{
	"":            true,
	"created_at":  true,
	"-created_at": true,
	"last_name":   true,
	"email":       true,
}

var ValidUserStatuses = map[string]bool{
	"active":    true,
	"suspended": true,
	"inactive":  true,
}

type UserListFilter struct {
	Search       string
	Role         string
	Status       string
	DepartmentID string
	Sort         string
	Limit        int
	Offset       int
}

type ListUsersQuery struct {
	Filter UserListFilter
}

type ListUsersResult struct {
	Users []adminout.UserSummary
	Total int
}

type ListUsersUseCase interface {
	Execute(ctx context.Context, query ListUsersQuery) (*ListUsersResult, error)
}

type GetUserUseCase interface {
	Execute(ctx context.Context, userID string) (*adminout.UserDetail, error)
}

type UpdateUserStatusCommand struct {
	UserID  string
	Status  string
	Reason  string
	ActorID string
}

type UpdateUserStatusUseCase interface {
	Execute(ctx context.Context, cmd UpdateUserStatusCommand) error
}

type UpdateUserRoleCommand struct {
	UserID  string
	Role    string
	Reason  string
	ActorID string
}

type UpdateUserRoleUseCase interface {
	Execute(ctx context.Context, cmd UpdateUserRoleCommand) error
}

type RevokeUserSessionsCommand struct {
	UserID  string
	Reason  string
	ActorID string
}

type RevokeUserSessionsUseCase interface {
	Execute(ctx context.Context, cmd RevokeUserSessionsCommand) error
}

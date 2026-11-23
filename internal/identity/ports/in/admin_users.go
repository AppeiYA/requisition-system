package identityin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

// AdminUsersPortIn groups the identity use cases used by the admin module.
// It is kept separate from IdentityPortIn so the existing constructor does not
// have to change for every new admin capability.
type AdminUsersPortIn struct {
	ListUsers         ListUsersUsecaseInt
	GetUser           GetUserUsecaseInt
	UpdateUserStatus  UpdateUserStatusUsecaseInt
	UpdateUserRole    UpdateUserRoleUsecaseInt
	RevokeUserSession RevokeUserSessionsUsecaseInt
}

type ListUsersResult struct {
	Users []domain.UserSummary
	Total int
}

type ListUsersUsecaseInt interface {
	Execute(ctx context.Context, filter domain.UserListFilter) (*ListUsersResult, error)
}

type GetUserUsecaseInt interface {
	Execute(ctx context.Context, userID string) (*domain.UserDetail, error)
}

type UpdateUserStatusCommand struct {
	UserID  string
	Status  string
	Reason  string
	ActorID string
}

type UpdateUserStatusUsecaseInt interface {
	Execute(ctx context.Context, cmd UpdateUserStatusCommand) error
}

type UpdateUserRoleCommand struct {
	UserID  string
	Role    string
	Reason  string
	ActorID string
}

type UpdateUserRoleUsecaseInt interface {
	Execute(ctx context.Context, cmd UpdateUserRoleCommand) error
}

type RevokeUserSessionsCommand struct {
	UserID  string
	Reason  string
	ActorID string
}

type RevokeUserSessionsUsecaseInt interface {
	Execute(ctx context.Context, cmd RevokeUserSessionsCommand) error
}

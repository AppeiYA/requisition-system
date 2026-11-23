package usecase

import (
	"context"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

type ListUsersUsecase struct{ users adminout.UserAdminPort }

func NewListUsersUsecase(users adminout.UserAdminPort) *ListUsersUsecase {
	return &ListUsersUsecase{users: users}
}

func (uc *ListUsersUsecase) Execute(ctx context.Context, query adminin.ListUsersQuery) (*adminin.ListUsersResult, error) {
	limit, offset := shareddomain.Clamp(query.Filter.Limit, query.Filter.Offset)
	filter := adminout.UserListFilter{
		Search: query.Filter.Search, Role: query.Filter.Role, Status: query.Filter.Status,
		DepartmentID: query.Filter.DepartmentID, Sort: query.Filter.Sort, Limit: limit, Offset: offset,
	}
	users, total, err := uc.users.ListUsers(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &adminin.ListUsersResult{Users: users, Total: total}, nil
}

package adminexternal

import (
	"context"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	identitydomain "github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
)

var _ adminout.UserAdminPort = (*AdminIdentityUserAdapter)(nil)

// AdminIdentityUserAdapter delegates admin user operations to the identity
// module's in-ports. All guards (self-change, last admin, activation) live in
// the identity use cases so they apply to every caller.
type AdminIdentityUserAdapter struct {
	users *identityin.AdminUsersPortIn
}

func NewAdminIdentityUserAdapter(users *identityin.AdminUsersPortIn) *AdminIdentityUserAdapter {
	return &AdminIdentityUserAdapter{users: users}
}

func (a *AdminIdentityUserAdapter) ListUsers(ctx context.Context, filter adminout.UserListFilter) ([]adminout.UserSummary, int, error) {
	result, err := a.users.ListUsers.Execute(ctx, identitydomain.UserListFilter{
		Search: filter.Search, Role: filter.Role, Status: filter.Status,
		DepartmentID: filter.DepartmentID, Sort: filter.Sort, Limit: filter.Limit, Offset: filter.Offset,
	})
	if err != nil {
		return nil, 0, err
	}
	users := make([]adminout.UserSummary, len(result.Users))
	for i, user := range result.Users {
		users[i] = toAdminUserSummary(user)
	}
	return users, result.Total, nil
}

func (a *AdminIdentityUserAdapter) GetUser(ctx context.Context, userID string) (*adminout.UserDetail, error) {
	user, err := a.users.GetUser.Execute(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := toAdminUserDetail(*user)
	return &result, nil
}

func toAdminUserSummary(user identitydomain.UserSummary) adminout.UserSummary {
	return adminout.UserSummary{
		ID: user.ID, FirstName: user.FirstName, LastName: user.LastName, Email: user.Email,
		StaffID: user.StaffID, Phone: user.Phone, Role: user.Role, Status: user.Status,
		DepartmentID: user.DepartmentID, DepartmentName: user.DepartmentName, JobTitle: user.JobTitle,
		MembershipStatus: user.MembershipStatus, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}
}

func toAdminUserDetail(user identitydomain.UserDetail) adminout.UserDetail {
	return adminout.UserDetail{
		UserSummary:       toAdminUserSummary(user.UserSummary),
		RequisitionCounts: user.RequisitionCounts, TotalEstimatedAmount: user.TotalEstimatedAmount,
		LastRequisitionAt: user.LastRequisitionAt,
	}
}

func (a *AdminIdentityUserAdapter) UpdateUserStatus(ctx context.Context, userID, status, reason, actorID string) error {
	return a.users.UpdateUserStatus.Execute(ctx, identityin.UpdateUserStatusCommand{
		UserID: userID, Status: status, Reason: reason, ActorID: actorID,
	})
}

func (a *AdminIdentityUserAdapter) UpdateUserRole(ctx context.Context, userID, role, reason, actorID string) error {
	return a.users.UpdateUserRole.Execute(ctx, identityin.UpdateUserRoleCommand{
		UserID: userID, Role: role, Reason: reason, ActorID: actorID,
	})
}

func (a *AdminIdentityUserAdapter) RevokeUserSessions(ctx context.Context, userID, reason, actorID string) error {
	return a.users.RevokeUserSession.Execute(ctx, identityin.RevokeUserSessionsCommand{
		UserID: userID, Reason: reason, ActorID: actorID,
	})
}

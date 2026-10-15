package invitationexternal

import (
	"context"

	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
)

var _ invitationout.InvitationDepartmentPort = (*InviteDepartmentAdapter)(nil)

// InviteDepartmentAdapter adapts the department module to the invitation
// module's outbound port. It delegates membership creation to the department
// module's in-port, so the invitation module never touches the department
// domain directly.
type InviteDepartmentAdapter struct {
	createMembership departmentin.CreateMembershipUsecaseInt
}

func NewInviteDepartmentAdapter(createMembership departmentin.CreateMembershipUsecaseInt) *InviteDepartmentAdapter {
	return &InviteDepartmentAdapter{createMembership: createMembership}
}

func (a *InviteDepartmentAdapter) CreateMembership(ctx context.Context, req *invitationout.CreateMembershipRequest) error {
	return a.createMembership.Execute(ctx, departmentin.CreateMembershipCommand{
		UserID:       req.UserID,
		DepartmentID: req.DepartmentID,
		JobTitle:     req.JobTitle,
	})
}

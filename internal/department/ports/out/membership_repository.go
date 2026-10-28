package departmentout

import (
	"context"

	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
)

type MembershipRepository interface {
	SaveMembership(ctx context.Context, membership *departmentdomain.DepartmentMembership) error
	GetMembershipByUserID(ctx context.Context, userID string) (*departmentdomain.DepartmentMembership, error)
}

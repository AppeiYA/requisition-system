package requisitionout

import (
	"context"

	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
)

type MembershipLookup interface {
	GetMembershipByUserID(ctx context.Context, userID string) (*departmentdomain.DepartmentMembership, error)
}

package departmentout

import (
	"context"

	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
)

type DepartmentRepository interface {
	SaveDepartment(ctx context.Context, department *departmentdomain.Department) error
	GetDepartmentByID(ctx context.Context, id string) (*departmentdomain.Department, error)
	GetDepartments(ctx context.Context) ([]*departmentdomain.Department, error)
}

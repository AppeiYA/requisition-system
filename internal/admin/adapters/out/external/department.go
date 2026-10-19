package adminexternal

import (
	"context"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
)

var _ adminout.AdminDepartmentPort = (*AdminDepartmentAdapter)(nil)

// AdminDepartmentAdapter adapts the department module to the admin module's
// outbound port. It resolves a department through the department module's
// in-port instead of reaching into the department domain directly.
type AdminDepartmentAdapter struct {
	getDepartmentByID departmentin.GetDepartmentByIDUsecaseInt
}

func NewAdminDepartmentAdapter(getDepartmentByID departmentin.GetDepartmentByIDUsecaseInt) *AdminDepartmentAdapter {
	return &AdminDepartmentAdapter{getDepartmentByID: getDepartmentByID}
}

func (a *AdminDepartmentAdapter) CheckDepartmentExists(ctx context.Context, departmentID string) (bool, error) {
	_, err := a.getDepartmentByID.Execute(ctx, departmentID)
	if err != nil {
		// A missing department is the expected "does not exist" answer.
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

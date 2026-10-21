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

// GetDepartmentName resolves the department's display name, which the
// invitation email quotes back to the invitee.
func (a *AdminDepartmentAdapter) GetDepartmentName(ctx context.Context, departmentID string) (string, error) {
	res, err := a.getDepartmentByID.Execute(ctx, departmentID)
	if err != nil {
		return "", err
	}
	return res.Name, nil
}

func (a *AdminDepartmentAdapter) GetDepartmentByID(ctx context.Context, departmentID string) (*adminout.DepartmentResponse, error) {
	res, err := a.getDepartmentByID.Execute(ctx, departmentID)
	if err != nil {
		return nil, err
	}
	
	return &adminout.DepartmentResponse{
		ID:          res.ID,
		Name:        res.Name,
		Description: res.Description,
		Code:        res.Code,
		Status:      res.Status,
	}, nil
}

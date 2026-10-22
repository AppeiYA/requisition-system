package adminout

import "context"

type DepartmentResponse struct {
	ID string
	Name string
	Description string
	Code string
	Status string
}
type AdminDepartmentPort interface {
	CheckDepartmentExists(ctx context.Context, departmentID string) (bool, error)
	GetDepartmentName(ctx context.Context, departmentID string) (string, error)
	GetDepartmentByID(ctx context.Context, departmentID string) (*DepartmentResponse, error)
}

package adminout

import "context"

// AdminDepartmentPort is the admin module's gateway to the department module.
// It lets an admin verify a department exists before inviting a member into
// it and resolve the department's display name for the invitation mail.
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

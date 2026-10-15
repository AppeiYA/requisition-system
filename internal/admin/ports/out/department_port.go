package adminout

import "context"

// AdminDepartmentPort is the admin module's gateway to the department module.
// It lets an admin verify a department exists before inviting a member into it.
type AdminDepartmentPort interface {
	CheckDepartmentExists(ctx context.Context, departmentID string) (bool, error)
}

package departmenthttp

// CreateDepartmentRequest is the admin's payload for creating a department.
type CreateDepartmentRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

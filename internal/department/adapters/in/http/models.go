package departmenthttp

import "github.com/AppeiYA/requisition-system/internal/department/domain"

// CreateDepartmentRequest is the admin's payload for creating a department.
type CreateDepartmentRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

type DepartmentResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Code        string `json:"code"`
	Status      string `json:"status"`
}

func DepartmentResponseFromDomain(dept *domain.Department) *DepartmentResponse {
	if dept == nil {
		return nil
	}

	return &DepartmentResponse{
		ID:          dept.ID(),
		Name:        dept.Name().Value(),
		Description: dept.Description().Value(),
		Code:        dept.Code().Value(),
		Status:      dept.Status().String(),
	}
}

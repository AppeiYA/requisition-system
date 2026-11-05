package adminhttp

type InviteRequest struct {
	Email        string `json:"email" binding:"required,email"`
	FirstName    string `json:"first_name" binding:"required"`
	LastName     string `json:"last_name" binding:"required"`
	Phone        string `json:"phone" binding:"omitempty"`
	StaffID      string `json:"staff_id" binding:"omitempty"`
	DepartmentID string `json:"department_id" binding:"required"`
	JobTitle     string `json:"job_title" binding:"required,max=60"`
}

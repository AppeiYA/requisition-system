package adminhttp

type InviteRequest struct {
	Email        string `json:"email" binding:"required,email"`
	FirstName    string `json:"firstName" binding:"required"`
	LastName     string `json:"lastName" binding:"required"`
	Phone        string `json:"phone" binding:"omitempty"`
	StaffID      string `json:"staffId" binding:"omitempty"`
	DepartmentID string `json:"departmentId" binding:"required"`
	JobTitle     string `json:"jobTitle" binding:"required,max=60"`
}

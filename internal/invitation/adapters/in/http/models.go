package invitationhttp

// AcceptInvitationRequest is the invitee's registration payload. First and last
// name come from the invitation's admin-provided prefill (the invitee cannot
// change them). Phone is optional and falls back to the prefilled value.
type AcceptInvitationRequest struct {
	Phone    string `json:"phone" binding:"omitempty"`
	Password string `json:"password" binding:"required"`
}

type ValidateInvitationResponse struct {
	Valid        bool   `json:"valid"`
	Email        string `json:"email,omitempty"`
	FirstName    string `json:"firstName,omitempty"`
	LastName     string `json:"lastName,omitempty"`
	Phone        string `json:"phone,omitempty"`
	StaffID      string `json:"staffId,omitempty"`
	DepartmentID string `json:"departmentId,omitempty"`
	JobTitle     string `json:"jobTitle,omitempty"`
}

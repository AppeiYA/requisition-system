package invitationhttp

type AcceptInvitationRequest struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

type ValidateInvitationResponse struct {
	Valid bool   `json:"valid"`
	Email string `json:"email,omitempty"`
}

package tasks

const SendInvitationEmail = "admin:send-invitation-email"

type InvitationEmailPayload struct {
	InvitationID   string `json:"invitation_id"`
	To             string `json:"to"`
	InvitedBy      string `json:"invited_by"`
	Token          string `json:"token"`
	ExpiresAt      string `json:"expires_at"`
	DepartmentName string `json:"department_name"`
	JobTitle       string `json:"job_title"`
	TraceParent    string `json:"trace_parent,omitempty"`
}

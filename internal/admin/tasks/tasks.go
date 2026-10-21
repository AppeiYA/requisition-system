package tasks

// SendInvitationEmail is the asynq task type for delivering an invitation
// email from the worker process.
const SendInvitationEmail = "admin:send-invitation-email"

// InvitationEmailPayload is the payload of a SendInvitationEmail task.
type InvitationEmailPayload struct {
	InvitationID   string `json:"invitation_id"`
	To             string `json:"to"`
	InvitedBy      string `json:"invited_by"`
	Token          string `json:"token"`
	ExpiresAt      string `json:"expires_at"`
	DepartmentName string `json:"department_name"`
	JobTitle       string `json:"job_title"`
}

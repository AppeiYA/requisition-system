package tasks

const (
	// SendPasswordResetEmail delivers one password reset link.
	SendPasswordResetEmail = "identity:send-password-reset-email"

	// CleanupPasswordResetTokens removes terminal reset rows past retention.
	CleanupPasswordResetTokens = "identity:cleanup-password-reset-tokens"
)

// PasswordResetEmailPayload carries everything the worker needs to deliver a
// reset link. The raw token lives only in the email and in this queue payload;
// it is never stored in PostgreSQL or put in a task ID.
type PasswordResetEmailPayload struct {
	ResetID     string `json:"reset_id"`
	To          string `json:"to"`
	Token       string `json:"token"`
	ExpiresAt   string `json:"expires_at"`
	TraceParent string `json:"trace_parent,omitempty"`
}

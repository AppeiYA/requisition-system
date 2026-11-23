package adminout

import (
	"context"
)

const OperationInviteMember = "admin.invitation.create"
const OperationSendInvitationEmail = "admin.email.delivery"
const OperationUpdateUserStatus = "admin.user.status.update"
const OperationUpdateUserRole = "admin.user.role.update"
const OperationRevokeUserSessions = "admin.user.sessions.revoke"
const OperationResendInvitation = "admin.invitation.resend"
const OperationRevokeInvitation = "admin.invitation.revoke"

type Telemetry interface {
	Measure(ctx context.Context, operation string, fn func(context.Context) error) error
}

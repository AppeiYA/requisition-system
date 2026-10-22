package adminout

import (
	"context"
)

const OperationInviteMember = "admin.invitation.create"
const OperationSendInvitationEmail = "admin.email.delivery"

type Telemetry interface {
	Measure(ctx context.Context, operation string, fn func(context.Context) error) error
}

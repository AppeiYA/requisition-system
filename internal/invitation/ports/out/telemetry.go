package invitationout

import (
	"context"
)

// OperationAcceptInvitation is the stable name of the invitation acceptance
// operation. It is reported as the span name and identifies the operation in
// the telemetry backend, so it must stay low cardinality and free of any
// request specific value such as the invitation token.
const OperationAcceptInvitation = "invitation.accept"

// OperationExpireInvitations is the stable name of the scheduled sweep that
// expires the invitations that were never accepted. It runs in the worker
// process, on a schedule, so it has no request to belong to and starts a
// trace of its own.
const OperationExpireInvitations = "invitation.expire"

// Telemetry measures the business operations of the invitation module. The
// port is deliberately free of tracing and metrics concepts: the application
// layer only asks for an operation to be measured and receives the outcome of
// the measured function unchanged.
type Telemetry interface {
	Measure(ctx context.Context, operation string, fn func(context.Context) error) error
}

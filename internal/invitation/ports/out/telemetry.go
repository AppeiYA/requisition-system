package invitationout

import (
	"context"
)

// OperationAcceptInvitation is the stable name of the invitation acceptance
// operation. It is reported as the span name and identifies the operation in
// the telemetry backend, so it must stay low cardinality and free of any
// request specific value such as the invitation token.
const OperationAcceptInvitation = "invitation.accept"

// Telemetry measures the business operations of the invitation module. The
// port is deliberately free of tracing and metrics concepts: the application
// layer only asks for an operation to be measured and receives the outcome of
// the measured function unchanged.
type Telemetry interface {
	Measure(ctx context.Context, operation string, fn func(context.Context) error) error
}

package identityout

import "context"

// Stable, low-cardinality operation names reported as spans and metrics. They
// must never contain a user, reset, or token identifier.
const (
	OperationRequestPasswordReset   = "auth.password_reset.request"
	OperationResetPassword          = "auth.password_reset.complete"
	OperationSendPasswordResetEmail = "auth.password_reset.email"
)

// Telemetry measures the business operations of the identity module. The port
// stays free of tracing and metrics concepts: the application layer only asks
// for an operation to be measured and receives the outcome unchanged.
type Telemetry interface {
	Measure(ctx context.Context, operation string, fn func(context.Context) error) error
}

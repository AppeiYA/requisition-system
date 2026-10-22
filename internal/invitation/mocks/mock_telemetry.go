package invitationmocks

import (
	"context"

	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
)

// TelemetryCall records a single measured operation.
type TelemetryCall struct {
	Operation string
	Err       error
}

// MockTelemetry records the measured operations. It behaves like the real
// adapter: the measured function is executed and its error is reported
// unchanged.
type MockTelemetry struct {
	MeasureFunc func(ctx context.Context, operation string, fn func(context.Context) error) error

	calls []TelemetryCall
}

func NewMockTelemetry() *MockTelemetry {
	return &MockTelemetry{
		MeasureFunc: func(ctx context.Context, operation string, fn func(context.Context) error) error {
			return fn(ctx)
		},
	}
}

func (m *MockTelemetry) Measure(ctx context.Context, operation string, fn func(context.Context) error) error {
	err := m.MeasureFunc(ctx, operation, fn)
	m.calls = append(m.calls, TelemetryCall{Operation: operation, Err: err})
	return err
}

// Calls returns the operations measured so far, in order.
func (m *MockTelemetry) Calls() []TelemetryCall {
	return append([]TelemetryCall(nil), m.calls...)
}

var _ invitationout.Telemetry = (*MockTelemetry)(nil)

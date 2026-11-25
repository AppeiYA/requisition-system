package identitymocks

import (
	"context"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

type MockTelemetry struct {
	MeasureFunc func(ctx context.Context, operation string, fn func(context.Context) error) error
	Operations  []string
}

func NewMockTelemetry() *MockTelemetry {
	return &MockTelemetry{
		MeasureFunc: func(ctx context.Context, operation string, fn func(context.Context) error) error {
			return fn(ctx)
		},
	}
}

func (m *MockTelemetry) Measure(ctx context.Context, operation string, fn func(context.Context) error) error {
	m.Operations = append(m.Operations, operation)
	return m.MeasureFunc(ctx, operation, fn)
}

var _ identityout.Telemetry = (*MockTelemetry)(nil)

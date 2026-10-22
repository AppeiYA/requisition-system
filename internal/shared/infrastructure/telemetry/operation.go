package telemetry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Operation names a measurable business operation. Name is the span name
// reported for the operation and MetricBase is the prefix of the metric names
// reported for it.
type Operation struct {
	Name       string
	MetricBase string
	Unit       string
}

// durationUnit is the unit of every duration reported by the adapter.
const durationUnit = "s"

type OperationTelemetry struct {
	operations map[string]*operationInstruments
}

type operationInstruments struct {
	attempted *Counter
	succeeded *Counter
	failed    *Counter
	duration  *Histogram
}

// NewOperationTelemetry builds the adapter and creates the instruments of
// every operation it is allowed to measure.
func NewOperationTelemetry(operations ...Operation) (*OperationTelemetry, error) {
	adapter := &OperationTelemetry{
		operations: make(map[string]*operationInstruments, len(operations)),
	}

	for _, operation := range operations {
		if operation.Name == "" {
			return nil, errors.New("telemetry: operation name is required")
		}
		if operation.MetricBase == "" {
			return nil, fmt.Errorf("telemetry: metric base is required for operation %q", operation.Name)
		}
		if _, exists := adapter.operations[operation.Name]; exists {
			return nil, fmt.Errorf("telemetry: operation %q is configured twice", operation.Name)
		}

		instruments, err := newOperationInstruments(operation)
		if err != nil {
			return nil, err
		}

		// The counters are published once with a zero value, so an outcome
		// that has not happened yet still has a series. A counter that was
		// never incremented has no series at all, and a panel cannot tell
		// "nothing failed" from "the metric is missing": the legend of the
		// outcome panels would silently leave the line out.
		instruments.publishZero(context.Background())

		adapter.operations[operation.Name] = instruments
	}

	return adapter, nil
}

// Measure runs fn and reports its outcome. The span is always ended and the
// error returned by fn is returned unchanged, so instrumentation can never
// replace nor mask the result of the business operation.
func (t *OperationTelemetry) Measure(ctx context.Context, operation string, fn func(context.Context) error) error {
	spanCtx, span := Start(ctx, operation)
	defer span.End()

	instruments, configured := t.operations[operation]
	if !configured {
		// The operation is still traced, but it has no instruments to report.
		AddEvent(span, "telemetry.operation_not_configured")
	} else {
		instruments.attempted.Inc(spanCtx)
	}

	start := time.Now()
	err := fn(spanCtx)
	elapsed := time.Since(start).Seconds()

	if configured {
		instruments.duration.Record(spanCtx, elapsed)
	}

	if err != nil {
		if configured {
			instruments.failed.Inc(spanCtx)
		}
		RecordError(span, err)
		return err
	}

	if configured {
		instruments.succeeded.Inc(spanCtx)
	}

	return nil
}

// publishZero records a zero on each counter of the operation.
//
// The point is published once, when the instruments are built, so the series
// exists from the start. The counters keep their own value once the operation
// runs: a zero adds nothing to a counter, it only gives it a series to be
// read from.
func (i *operationInstruments) publishZero(ctx context.Context) {
	i.attempted.Add(ctx, 0)
	i.succeeded.Add(ctx, 0)
	i.failed.Add(ctx, 0)
}

func newOperationInstruments(operation Operation) (*operationInstruments, error) {
	attempted, err := NewCounter(
		operation.MetricBase+".attempted",
		"Number of attempts of "+operation.Name,
		operation.Unit,
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: cannot create the attempted counter of %q: %w", operation.Name, err)
	}

	succeeded, err := NewCounter(
		operation.MetricBase+".succeeded",
		"Number of successful "+operation.Name+" operations",
		operation.Unit,
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: cannot create the succeeded counter of %q: %w", operation.Name, err)
	}

	failed, err := NewCounter(
		operation.MetricBase+".failed",
		"Number of failed "+operation.Name+" operations",
		operation.Unit,
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: cannot create the failed counter of %q: %w", operation.Name, err)
	}

	duration, err := NewHistogram(
		operation.MetricBase+".duration",
		"Duration of "+operation.Name+" operations in seconds",
		durationUnit,
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: cannot create the duration histogram of %q: %w", operation.Name, err)
	}

	return &operationInstruments{
		attempted: attempted,
		succeeded: succeeded,
		failed:    failed,
		duration:  duration,
	}, nil
}

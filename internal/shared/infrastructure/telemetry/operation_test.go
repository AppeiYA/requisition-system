package telemetry

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

const (
	testOperation  = "admin.invitation.create"
	testMetricBase = "requisition.admin.invitations"
)

// setupTelemetry installs a span recorder and a manual metric reader as the
// global providers, so the measurements can be asserted without relying on
// the stdout exporters.
func setupTelemetry(t *testing.T) (*tracetest.SpanRecorder, *sdkmetric.ManualReader) {
	t.Helper()

	recorder := tracetest.NewSpanRecorder()
	traceProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(traceProvider)

	reader := sdkmetric.NewManualReader()
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(meterProvider)

	t.Cleanup(func() {
		_ = traceProvider.Shutdown(context.Background())
		_ = meterProvider.Shutdown(context.Background())
	})

	return recorder, reader
}

func newTestTelemetry(t *testing.T) *OperationTelemetry {
	t.Helper()

	adapter, err := NewOperationTelemetry(Operation{Name: testOperation, MetricBase: testMetricBase})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	return adapter
}

func collect(t *testing.T, reader *sdkmetric.ManualReader) metricdata.ResourceMetrics {
	t.Helper()

	var metrics metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &metrics); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	return metrics
}

func counterValue(t *testing.T, metrics metricdata.ResourceMetrics, name string) int64 {
	t.Helper()

	for _, scope := range metrics.ScopeMetrics {
		for _, metric := range scope.Metrics {
			if metric.Name != name {
				continue
			}

			sum, ok := metric.Data.(metricdata.Sum[int64])
			if !ok {
				t.Fatalf("expected %q to be a sum of int64, got %T", name, metric.Data)
			}

			var total int64
			for _, point := range sum.DataPoints {
				total += point.Value
			}

			return total
		}
	}

	// An instrument without any recorded value is not reported at all.
	return 0
}

func histogramCount(t *testing.T, metrics metricdata.ResourceMetrics, name string) uint64 {
	t.Helper()

	for _, scope := range metrics.ScopeMetrics {
		for _, metric := range scope.Metrics {
			if metric.Name != name {
				continue
			}

			histogram, ok := metric.Data.(metricdata.Histogram[float64])
			if !ok {
				t.Fatalf("expected %q to be a histogram of float64, got %T", name, metric.Data)
			}

			var count uint64
			for _, point := range histogram.DataPoints {
				count += point.Count
			}

			return count
		}
	}

	return 0
}

func Test_NewOperationTelemetry_InvalidOperations(t *testing.T) {
	tests := []struct {
		name       string
		operations []Operation
	}{
		{
			name:       "missing name",
			operations: []Operation{{MetricBase: testMetricBase}},
		},
		{
			name:       "missing metric base",
			operations: []Operation{{Name: testOperation}},
		},
		{
			name: "duplicated operation",
			operations: []Operation{
				{Name: testOperation, MetricBase: testMetricBase},
				{Name: testOperation, MetricBase: testMetricBase},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewOperationTelemetry(tt.operations...); err == nil {
				t.Fatal("expected an error, got none")
			}
		})
	}
}

func Test_OperationTelemetry_MeasureSuccess(t *testing.T) {
	recorder, reader := setupTelemetry(t)
	adapter := newTestTelemetry(t)

	err := adapter.Measure(context.Background(), testOperation, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Name() != testOperation {
		t.Errorf("expected span %q, got %q", testOperation, spans[0].Name())
	}
	if spans[0].Status().Code != codes.Unset {
		t.Errorf("expected the span to stay unset, got %q", spans[0].Status().Code)
	}

	metrics := collect(t, reader)
	if got := counterValue(t, metrics, testMetricBase+".attempted"); got != 1 {
		t.Errorf("expected 1 attempt, got %d", got)
	}
	if got := counterValue(t, metrics, testMetricBase+".succeeded"); got != 1 {
		t.Errorf("expected 1 success, got %d", got)
	}
	if got := counterValue(t, metrics, testMetricBase+".failed"); got != 0 {
		t.Errorf("expected no failure, got %d", got)
	}
	if got := histogramCount(t, metrics, testMetricBase+".duration"); got != 1 {
		t.Errorf("expected 1 duration measurement, got %d", got)
	}
}

func Test_OperationTelemetry_MeasureFailure(t *testing.T) {
	recorder, reader := setupTelemetry(t)
	adapter := newTestTelemetry(t)

	operationErr := errors.New("invitation could not be saved")

	err := adapter.Measure(context.Background(), testOperation, func(ctx context.Context) error {
		return operationErr
	})
	if !errors.Is(err, operationErr) {
		t.Fatalf("expected operationErr, got: %v", err)
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status().Code != codes.Error {
		t.Errorf("expected the span to be in error, got %q", spans[0].Status().Code)
	}
	if len(spans[0].Events()) == 0 {
		t.Error("expected the error to be recorded on the span")
	}

	metrics := collect(t, reader)
	if got := counterValue(t, metrics, testMetricBase+".attempted"); got != 1 {
		t.Errorf("expected 1 attempt, got %d", got)
	}
	if got := counterValue(t, metrics, testMetricBase+".failed"); got != 1 {
		t.Errorf("expected 1 failure, got %d", got)
	}
	if got := counterValue(t, metrics, testMetricBase+".succeeded"); got != 0 {
		t.Errorf("expected no success, got %d", got)
	}
	if got := histogramCount(t, metrics, testMetricBase+".duration"); got != 1 {
		t.Errorf("expected 1 duration measurement, got %d", got)
	}
}

func Test_OperationTelemetry_MeasurePropagatesTheTraceContext(t *testing.T) {
	recorder, _ := setupTelemetry(t)
	adapter := newTestTelemetry(t)

	var spanContext trace.SpanContext
	err := adapter.Measure(context.Background(), testOperation, func(ctx context.Context) error {
		spanContext = trace.SpanContextFromContext(ctx)
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if !spanContext.IsValid() {
		t.Fatal("expected the measured operation to run inside the span")
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].SpanContext().SpanID() != spanContext.SpanID() {
		t.Errorf("expected the measured operation to run inside the recorded span")
	}
}

func Test_OperationTelemetry_MeasureUnknownOperation(t *testing.T) {
	recorder, reader := setupTelemetry(t)
	adapter := newTestTelemetry(t)

	operationErr := errors.New("department not found")

	err := adapter.Measure(context.Background(), "admin.unknown", func(ctx context.Context) error {
		return operationErr
	})
	if !errors.Is(err, operationErr) {
		t.Fatalf("expected operationErr, got: %v", err)
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Name() != "admin.unknown" {
		t.Errorf("expected span %q, got %q", "admin.unknown", spans[0].Name())
	}
	if len(spans[0].Events()) == 0 {
		t.Error("expected the unconfigured operation to be reported on the span")
	}

	metrics := collect(t, reader)
	if got := counterValue(t, metrics, testMetricBase+".attempted"); got != 0 {
		t.Errorf("expected no attempt, got %d", got)
	}
}

package taskqueue

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/telemetry"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// recordingTracer installs a tracer provider that keeps every span, so a test
// can look at what the middleware reported. The middleware reads the global
// provider, so it has to be installed before it builds the handler, exactly
// like in the worker process.
func recordingTracer(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()

	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))

	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(sdktrace.NewTracerProvider())
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown: %v", err)
		}
	})

	return recorder
}

// onlySpan is the span the middleware started, told apart from the ones the
// test itself creates.
func onlySpan(t *testing.T, recorder *tracetest.SpanRecorder) sdktrace.ReadOnlySpan {
	t.Helper()

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("ended spans = %d, want 1", len(spans))
	}
	return spans[0]
}

// Test_TraceMiddleware_ContinuesTheTraceOfTheProducer checks the whole point
// of the task payload carrying a traceparent: the work of the worker belongs
// to the trace of the request that queued the task.
func Test_TraceMiddleware_ContinuesTheTraceOfTheProducer(t *testing.T) {
	recorder := recordingTracer(t)

	provider := sdktrace.NewTracerProvider()
	tracer := provider.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "POST /api/admin/invitations")
	defer span.End()

	task := asynq.NewTask("admin:send-invitation-email", traceParentedPayload(t, telemetry.TraceParentFrom(ctx)))

	var handlerSpanContext trace.SpanContext
	handler := TraceMiddleware()(asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
		handlerSpanContext = trace.SpanContextFromContext(ctx)
		return nil
	}))

	if err := handler.ProcessTask(context.Background(), task); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if !handlerSpanContext.IsValid() {
		t.Fatal("expected the handler to run inside a span")
	}
	if handlerSpanContext.TraceID() != span.SpanContext().TraceID() {
		t.Errorf("trace ID = %v, want %v", handlerSpanContext.TraceID(), span.SpanContext().TraceID())
	}

	got := onlySpan(t, recorder)
	if got.SpanContext().TraceID() != span.SpanContext().TraceID() {
		t.Errorf("span trace ID = %v, want %v", got.SpanContext().TraceID(), span.SpanContext().TraceID())
	}
	if kind := got.SpanKind(); kind != trace.SpanKindConsumer {
		t.Errorf("span kind = %v, want %v", kind, trace.SpanKindConsumer)
	}
}

// Test_TraceMiddleware_WithoutAParentStartsARootTrace checks the scheduled
// tasks, whose payload carries nothing, are still traced: they have no
// request to belong to, so they start a trace of their own.
func Test_TraceMiddleware_WithoutAParentStartsARootTrace(t *testing.T) {
	recorder := recordingTracer(t)

	handler := TraceMiddleware()(asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
		return nil
	}))

	// A scheduled task carries no payload at all.
	if err := handler.ProcessTask(context.Background(), asynq.NewTask("invitation:expire", nil)); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	got := onlySpan(t, recorder)
	if got.Name() != "asynq invitation:expire" {
		t.Errorf("span name = %q, want %q", got.Name(), "asynq invitation:expire")
	}
	if got.Parent().IsValid() {
		t.Error("expected a root span, got a child of another span")
	}
}

// Test_TraceMiddleware_APayloadThatIsNotJSONHasNoParent checks the one field
// the middleware understands is looked for without assuming the payload is
// well formed: a payload it cannot read simply starts a new trace.
func Test_TraceMiddleware_APayloadThatIsNotJSONHasNoParent(t *testing.T) {
	recorder := recordingTracer(t)

	handler := TraceMiddleware()(asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
		return nil
	}))

	if err := handler.ProcessTask(context.Background(), asynq.NewTask("admin:send-invitation-email", []byte{0x01, 0x02})); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if got := onlySpan(t, recorder); got.Parent().IsValid() {
		t.Error("expected a root span, got a child of another span")
	}
}

// Test_TraceMiddleware_RecordsTheFailure checks a task that gave up is
// reported on its span, and that the error reaches asynq untouched so the
// retry and archive rules still apply.
func Test_TraceMiddleware_RecordsTheFailure(t *testing.T) {
	recorder := recordingTracer(t)

	sendErr := errors.New("send failed")
	handler := TraceMiddleware()(asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
		return sendErr
	}))

	err := handler.ProcessTask(context.Background(), asynq.NewTask("admin:send-invitation-email", nil))
	if !errors.Is(err, sendErr) {
		t.Fatalf("expected the handler error unchanged, got: %v", err)
	}

	got := onlySpan(t, recorder)
	if got.Status().Code != codes.Error {
		t.Errorf("span status = %v, want %v", got.Status().Code, codes.Error)
	}
	if len(got.Events()) == 0 {
		t.Error("expected the failure to be recorded as a span event")
	}
}

func traceParentedPayload(t *testing.T, traceParent string) []byte {
	t.Helper()

	b, err := json.Marshal(map[string]string{"trace_parent": traceParent})
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	return b
}

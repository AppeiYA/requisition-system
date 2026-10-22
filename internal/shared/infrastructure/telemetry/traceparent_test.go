package telemetry

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Test_TraceParent_RoundTrip checks what one process writes into a task
// payload is what the other one reads back, so the trace of the request that
// queued the task carries on into the worker that runs it.
func Test_TraceParent_RoundTrip(t *testing.T) {
	t.Parallel()

	provider := sdktrace.NewTracerProvider()
	ctx, span := provider.Tracer("test").Start(context.Background(), "POST /api/admin/invitations")
	defer span.End()

	traceParent := TraceParentFrom(ctx)
	if traceParent == "" {
		t.Fatal("expected a traceparent for a context carrying a span")
	}

	parentCtx, ok := TraceParentContext(context.Background(), traceParent)
	if !ok {
		t.Fatal("expected the traceparent to name a usable parent")
	}

	got := trace.SpanContextFromContext(parentCtx)
	if got.TraceID() != span.SpanContext().TraceID() {
		t.Errorf("trace ID = %v, want %v", got.TraceID(), span.SpanContext().TraceID())
	}
	if got.SpanID() != span.SpanContext().SpanID() {
		t.Errorf("span ID = %v, want %v", got.SpanID(), span.SpanContext().SpanID())
	}
	if !got.IsRemote() {
		t.Error("expected the parent to be marked remote: it ran in another process")
	}
}

// Test_TraceParent_FromAContextWithoutASpan checks a process with nothing to
// hand over leaves the payload free of the field instead of writing a
// meaningless one.
func Test_TraceParent_FromAContextWithoutASpan(t *testing.T) {
	t.Parallel()

	if got := TraceParentFrom(context.Background()); got != "" {
		t.Errorf("TraceParentFrom() = %q, want an empty value", got)
	}

	parentCtx, ok := TraceParentContext(context.Background(), "")
	if ok {
		t.Error("expected no parent for an empty traceparent")
	}
	if trace.SpanContextFromContext(parentCtx).IsValid() {
		t.Error("expected the context to be left unchanged")
	}
}

// Test_TraceParent_FromAnUnusableTraceParent checks a value that does not
// name a span is told apart from a real parent, so a corrupted payload
// starts a new trace instead of attaching the task to nothing.
func Test_TraceParent_FromAnUnusableTraceParent(t *testing.T) {
	t.Parallel()

	for _, traceParent := range []string{
		"not-a-traceparent",
		"00-",
		"00-00000000000000000000000000000000-0000000000000000-01",
	} {
		if _, ok := TraceParentContext(context.Background(), traceParent); ok {
			t.Errorf("expected no parent for %q", traceParent)
		}
	}
}

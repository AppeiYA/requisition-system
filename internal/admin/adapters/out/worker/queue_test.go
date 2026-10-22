package adminworker

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/admin/tasks"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/telemetry"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func Test_InvitationEmailTask_RoundTrip(t *testing.T) {
	t.Parallel()

	payload := tasks.InvitationEmailPayload{
		To:        "invitee@example.com",
		InvitedBy: "usr-1234",
		Token:     "raw-invitation-token",
		ExpiresAt: "2026-01-01T00:00:00Z",
	}

	task, err := invitationEmailTask(context.Background(), payload)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if task.Type() != tasks.SendInvitationEmail {
		t.Errorf("expected task type %q, got %q", tasks.SendInvitationEmail, task.Type())
	}

	var decoded tasks.InvitationEmailPayload
	if err := json.Unmarshal(task.Payload(), &decoded); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}
	if decoded != payload {
		t.Errorf("payload did not round-trip: got %+v, want %+v", decoded, payload)
	}
}

// Test_InvitationEmailTask_CarriesTheTraceOfTheQueuingRequest checks the
// worker can pick the trace up where the API process leaves it: the payload
// holds the traceparent of the context the task was queued in, and it names
// a span the worker can continue.
func Test_InvitationEmailTask_CarriesTheTraceOfTheQueuingRequest(t *testing.T) {
	t.Parallel()

	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))

	// The span context of the queuing request is what travels, so it does not
	// matter which provider is installed: only ctx is read here.
	ctx, span := provider.Tracer("test").Start(context.Background(), "POST /api/admin/invitations")
	defer span.End()

	task, err := invitationEmailTask(ctx, tasks.InvitationEmailPayload{Token: "raw-invitation-token"})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	var decoded tasks.InvitationEmailPayload
	if err := json.Unmarshal(task.Payload(), &decoded); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	parentCtx, ok := telemetry.TraceParentContext(context.Background(), decoded.TraceParent)
	if !ok {
		t.Fatal("expected the traceparent of the payload to name a usable parent")
	}
	if got := trace.SpanContextFromContext(parentCtx).TraceID(); got != span.SpanContext().TraceID() {
		t.Errorf("trace ID = %v, want %v", got, span.SpanContext().TraceID())
	}
}

// Test_InvitationEmailTask_QueuedOutsideASpanHasNoTraceParent checks the
// payload stays free of the field when there is no span to hand over, so a
// task queued by a process with no telemetry still decodes the same way.
func Test_InvitationEmailTask_QueuedOutsideASpanHasNoTraceParent(t *testing.T) {
	t.Parallel()

	task, err := invitationEmailTask(context.Background(), tasks.InvitationEmailPayload{Token: "raw-invitation-token"})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	var decoded tasks.InvitationEmailPayload
	if err := json.Unmarshal(task.Payload(), &decoded); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}
	if decoded.TraceParent != "" {
		t.Errorf("trace_parent = %q, want an empty value", decoded.TraceParent)
	}
}

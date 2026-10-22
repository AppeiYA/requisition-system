package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// instrumentationName identifies this application in the telemetry backend.
const instrumentationName = "requisition-system"

// tracer is resolved on every call so the span always goes to the provider
// that is active at that moment, including the one installed after start up.
func tracer() trace.Tracer {
	return otel.Tracer(instrumentationName)
}

func Start(ctx context.Context, operation string) (context.Context, trace.Span) {
	return tracer().Start(ctx, operation)
}

// StartConsumer starts the span of work handed over by another process: a
// queued task for instance. traceParent is the W3C traceparent the producer
// left behind, and when it names a usable span the task carries on the trace
// of the request that queued it. A task with no parent, the scheduled ones
// for instance, starts a trace of its own.
func StartConsumer(ctx context.Context, name, traceParent string) (context.Context, trace.Span) {
	ctx, _ = TraceParentContext(ctx, traceParent)
	return tracer().Start(ctx, name, trace.WithSpanKind(trace.SpanKindConsumer))
}

func RecordError(span trace.Span, err error) {
	if err == nil {
		return
	}

	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

func SetAttribute(
	span trace.Span,
	key string,
	value string,
) {
	span.SetAttributes(
		attribute.String(key, value),
	)
}

// AddEvent records a named event on the span without changing its status.
func AddEvent(span trace.Span, name string) {
	span.AddEvent(name)
}

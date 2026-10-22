package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// traceParentHeader is the W3C propagation key. The queued task payloads
// carry the very same name as a JSON field, so a task travels with the trace
// of the request that queued it.
const traceParentHeader = "traceparent"

// TraceParentFrom reports the W3C traceparent of ctx, or an empty string when
// ctx carries no span worth handing to another process.
//
// The W3C propagator is named directly instead of the configured one: the
// value is stored in a single field whose name fixes the format, and baggage
// has nowhere to ride along.
func TraceParentFrom(ctx context.Context) string {
	carrier := propagation.MapCarrier{}
	propagation.TraceContext{}.Inject(ctx, carrier)
	return carrier.Get(traceParentHeader)
}

// TraceParentContext continues ctx with the span that traceParent names, so
// the work of another process becomes part of the trace it started. The
// boolean reports whether a usable parent was found; when it is false, ctx is
// returned unchanged and the caller starts a new trace of its own.
func TraceParentContext(ctx context.Context, traceParent string) (context.Context, bool) {
	if traceParent == "" {
		return ctx, false
	}

	parent := propagation.TraceContext{}.Extract(ctx, propagation.MapCarrier{
		traceParentHeader: traceParent,
	})

	// A traceparent that does not parse leaves no span context behind, which
	// is how a corrupted or foreign value is told apart from a real parent.
	if trace.SpanContextFromContext(parent).IsValid() {
		return parent, true
	}
	return ctx, false
}

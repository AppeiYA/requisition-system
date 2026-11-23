package logger

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// ContextHandler is an slog.Handler middleware that enriches log records
// with OpenTelemetry trace_id, span_id, and trace_flags when an active
// trace exists in the context.
type ContextHandler struct {
	handler slog.Handler
}

// NewContextHandler wraps an existing slog.Handler with OpenTelemetry context enrichment.
func NewContextHandler(handler slog.Handler) *ContextHandler {
	return &ContextHandler{
		handler: handler,
	}
}

var _ slog.Handler = (*ContextHandler)(nil)

func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx == nil {
		return h.handler.Handle(ctx, r)
	}

	span := trace.SpanFromContext(ctx)
	if span != nil && span.SpanContext().IsValid() {
		sc := span.SpanContext()
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
			slog.String("trace_flags", sc.TraceFlags().String()),
		)
	}

	return h.handler.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{
		handler: h.handler.WithAttrs(attrs),
	}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{
		handler: h.handler.WithGroup(name),
	}
}

package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("requisition-system")

func Start(ctx context.Context, operation string) (context.Context, trace.Span) {
	return tracer.Start(ctx, operation)
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

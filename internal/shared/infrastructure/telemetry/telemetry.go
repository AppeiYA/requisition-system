package telemetry

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Telemetry struct {
	serviceName string
}

func NewTelemetry(serviceName string) *Telemetry {
	return &Telemetry{
		serviceName: serviceName,
	}
}

func (t *Telemetry) Setup(ctx context.Context) (func(ctx context.Context) error, error) {
	return setup(ctx, t.serviceName)
}

func setup(ctx context.Context, serviceName string) (func(ctx context.Context) error, error) {
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
		),
	)

	if err != nil {
		return nil, err
	}

	traceExporter, err := stdouttrace.New()
	if err != nil {
		return nil, err
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter),
	)

	metricExporter, err := stdoutmetric.New()
	if err != nil {
		_ = traceProvider.Shutdown(ctx)
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(metricExporter),
		),
	)

	otel.SetTracerProvider(traceProvider)
	otel.SetMeterProvider(meterProvider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	shutdown := func(ctx context.Context) error {
		return errors.Join(
			traceProvider.Shutdown(ctx),
			meterProvider.Shutdown(ctx),
		)
	}

	return shutdown, nil
}

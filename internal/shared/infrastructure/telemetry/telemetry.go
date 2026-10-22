package telemetry

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/version"
)

// ErrExporterUnsupported is returned when the configured exporter is not one
// of the supported ones.
var ErrExporterUnsupported = errors.New("telemetry: unsupported exporter")

// ErrProtocolUnsupported is returned when the configured OTLP protocol is not
// one of the supported ones.
var ErrProtocolUnsupported = errors.New("telemetry: unsupported OTLP protocol")

// Telemetry owns the OpenTelemetry pipeline of a process. It builds the
// resource that describes the process, the trace and meter providers, and
// installs them as the global providers.
//
// The exporters are chosen by the configuration: stdout keeps the telemetry on
// the terminal, which is what a developer without a collector gets, and otlp
// ships it to a collector that forwards it to the storage backends.
type Telemetry struct {
	serviceName string
	cfg         config.TelemetryConfig
}

// NewTelemetry builds the telemetry of a process. The service name identifies
// the process in the backend; the configuration says where the telemetry
// goes. The service name is only a default: OTEL_SERVICE_NAME overrides it.
func NewTelemetry(serviceName string, cfg config.TelemetryConfig) *Telemetry {
	return &Telemetry{
		serviceName: serviceName,
		cfg:         cfg,
	}
}

// ServiceName is the name the backend knows the process by: the configured
// one when it overrides the built-in default, the default otherwise. The
// instrumented clients that take a service name of their own ask for it
// here, so every instrument agrees on the identity of the process.
func (t *Telemetry) ServiceName() string {
	if t.cfg.ServiceName != "" {
		return t.cfg.ServiceName
	}
	return t.serviceName
}

// Setup installs the global tracer and meter providers and returns the
// function that flushes and closes them.
//
// It must be called before the instrumented clients are created, so the
// clients pick up the providers that are installed here.
func (t *Telemetry) Setup(ctx context.Context) (func(ctx context.Context) error, error) {
	res, err := t.resource(ctx)
	if err != nil {
		return nil, err
	}

	sampler, err := newSampler(t.cfg.Sampler, t.cfg.SampleRatio)
	if err != nil {
		return nil, err
	}

	traceProvider, err := t.newTraceProvider(ctx, res, sampler)
	if err != nil {
		return nil, err
	}

	meterProvider, err := t.newMeterProvider(ctx, res)
	if err != nil {
		return nil, errors.Join(err, traceProvider.Shutdown(ctx))
	}

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

// resource describes the process the telemetry comes from.
//
// The attributes declared in code are the defaults and the ones read from the
// environment win, so a deployment can override them without a rebuild.
func (t *Telemetry) resource(ctx context.Context) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(t.ServiceName()),
		semconv.ServiceVersion(version.Version),
		semconv.ServiceInstanceID(instanceID()),
	}

	if t.cfg.Environment != "" {
		attrs = append(attrs, semconv.DeploymentEnvironmentNameKey.String(t.cfg.Environment))
	}

	return resource.New(ctx,
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithProcess(),
		resource.WithAttributes(attrs...),
		// The environment is read last, so OTEL_RESOURCE_ATTRIBUTES and
		// OTEL_SERVICE_NAME override the defaults above.
		resource.WithFromEnv(),
	)
}

// instanceID tells the instances of the same service apart. It is generated
// once per process, so the traces of two replicas never mix.
func instanceID() string {
	return uuid.NewString()
}

func (t *Telemetry) newTraceProvider(
	ctx context.Context,
	res *resource.Resource,
	sampler sdktrace.Sampler,
) (*sdktrace.TracerProvider, error) {
	switch t.cfg.Exporter {
	case config.TelemetryExporterNone:
		// Nothing leaves the process, so nothing is sampled either.
		return sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.NeverSample()),
		), nil

	case config.TelemetryExporterStdout:
		exporter, err := stdouttrace.New()
		if err != nil {
			return nil, fmt.Errorf("telemetry: create the stdout trace exporter: %w", err)
		}
		return sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sampler),
			sdktrace.WithBatcher(exporter),
		), nil

	case config.TelemetryExporterOTLP:
		exporter, err := t.newOTLPTraceExporter(ctx)
		if err != nil {
			return nil, err
		}
		return sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sampler),
			sdktrace.WithBatcher(exporter),
		), nil

	default:
		return nil, fmt.Errorf("%w: %q", ErrExporterUnsupported, t.cfg.Exporter)
	}
}

func (t *Telemetry) newMeterProvider(
	ctx context.Context,
	res *resource.Resource,
) (*sdkmetric.MeterProvider, error) {
	switch t.cfg.Exporter {
	case config.TelemetryExporterNone:
		// No reader, so nothing is collected.
		return sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
		), nil

	case config.TelemetryExporterStdout:
		exporter, err := stdoutmetric.New()
		if err != nil {
			return nil, fmt.Errorf("telemetry: create the stdout metric exporter: %w", err)
		}
		return sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(
				sdkmetric.NewPeriodicReader(
					exporter,
					sdkmetric.WithInterval(t.cfg.MetricsInterval),
				),
			),
		), nil

	case config.TelemetryExporterOTLP:
		exporter, err := t.newOTLPMetricExporter(ctx)
		if err != nil {
			return nil, err
		}
		return sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(
				sdkmetric.NewPeriodicReader(
					exporter,
					sdkmetric.WithInterval(t.cfg.MetricsInterval),
				),
			),
		), nil

	default:
		return nil, fmt.Errorf("%w: %q", ErrExporterUnsupported, t.cfg.Exporter)
	}
}

func (t *Telemetry) newOTLPTraceExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	switch t.cfg.OTLPProtocol {
	case config.OTLPProtocolHTTP:
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(t.cfg.OTLPEndpoint),
		}
		if t.cfg.OTLPInsecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}

		exporter, err := otlptracehttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("telemetry: create the OTLP trace exporter: %w", err)
		}
		return exporter, nil

	case config.OTLPProtocolGRPC:
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(t.cfg.OTLPEndpoint),
		}
		if t.cfg.OTLPInsecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}

		exporter, err := otlptracegrpc.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("telemetry: create the OTLP trace exporter: %w", err)
		}
		return exporter, nil

	default:
		return nil, fmt.Errorf("%w: %q", ErrProtocolUnsupported, t.cfg.OTLPProtocol)
	}
}

func (t *Telemetry) newOTLPMetricExporter(ctx context.Context) (sdkmetric.Exporter, error) {
	switch t.cfg.OTLPProtocol {
	case config.OTLPProtocolHTTP:
		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(t.cfg.OTLPEndpoint),
		}
		if t.cfg.OTLPInsecure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}

		exporter, err := otlpmetrichttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("telemetry: create the OTLP metric exporter: %w", err)
		}
		return exporter, nil

	case config.OTLPProtocolGRPC:
		opts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(t.cfg.OTLPEndpoint),
		}
		if t.cfg.OTLPInsecure {
			opts = append(opts, otlpmetricgrpc.WithInsecure())
		}

		exporter, err := otlpmetricgrpc.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("telemetry: create the OTLP metric exporter: %w", err)
		}
		return exporter, nil

	default:
		return nil, fmt.Errorf("%w: %q", ErrProtocolUnsupported, t.cfg.OTLPProtocol)
	}
}

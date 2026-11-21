package config

import (
	"fmt"
	"time"
)

const (
	TelemetryExporterStdout = "stdout"
	TelemetryExporterOTLP   = "otlp"
	TelemetryExporterNone   = "none"
)

const (
	OTLPProtocolGRPC = "grpc"
	OTLPProtocolHTTP = "http"
)

// TelemetryConfig configures how the OpenTelemetry traces and metrics leave
// the process. The variable names follow the OpenTelemetry specification, so
// the application can be pointed at any OTLP backend without a code change.
type TelemetryConfig struct {
	ServiceName     string
	Environment     string
	Exporter        string
	OTLPEndpoint    string
	OTLPProtocol    string
	OTLPInsecure    bool
	Sampler         string
	SampleRatio     float64
	MetricsInterval time.Duration
}

func LoadTelemetryConfig() (*TelemetryConfig, error) {
	exporter := getEnv("TELEMETRY_EXPORTER", TelemetryExporterStdout)
	switch exporter {
	case TelemetryExporterStdout, TelemetryExporterOTLP, TelemetryExporterNone:
	default:
		return nil, fmt.Errorf("%w, got %q", ErrTelemetryExporterUnsupported, exporter)
	}

	protocol := getEnv("OTEL_EXPORTER_OTLP_PROTOCOL", OTLPProtocolGRPC)
	switch protocol {
	case OTLPProtocolGRPC, OTLPProtocolHTTP:
	default:
		return nil, fmt.Errorf("%w, got %q", ErrTelemetryProtocolUnsupported, protocol)
	}

	sampler := getEnv("OTEL_TRACES_SAMPLER", "parentbased_traceidratio")

	sampleRatio, err := getFloatEnv("OTEL_TRACES_SAMPLER_ARG", 1)
	if err != nil {
		return nil, err
	}
	if sampleRatio < 0 || sampleRatio > 1 {
		return nil, fmt.Errorf("%w, got %v", ErrTelemetrySampleRatioOutOfRange, sampleRatio)
	}

	otlpInsecure, err := getBoolEnv("OTEL_EXPORTER_OTLP_INSECURE", true)
	if err != nil {
		return nil, err
	}

	// The OpenTelemetry specification counts this interval in milliseconds,
	// and the SDK reads the same variable: the two have to agree on the unit
	// or the SDK logs a parse error on every start.
	metricsInterval, err := getDurationMillisEnv("OTEL_METRIC_EXPORT_INTERVAL", 60*time.Second)
	if err != nil {
		return nil, err
	}
	if metricsInterval <= 0 {
		return nil, fmt.Errorf("%w, got %v", ErrTelemetryMetricsIntervalInvalid, metricsInterval)
	}

	return &TelemetryConfig{
		ServiceName:     getEnv("OTEL_SERVICE_NAME", ""),
		Environment:     getEnv("APP_ENV", "development"),
		Exporter:        exporter,
		OTLPEndpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		OTLPProtocol:    protocol,
		OTLPInsecure:    otlpInsecure,
		Sampler:         sampler,
		SampleRatio:     sampleRatio,
		MetricsInterval: metricsInterval,
	}, nil
}
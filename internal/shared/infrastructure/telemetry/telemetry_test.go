package telemetry

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/sdk/resource"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/version"
)

// baseConfig returns a configuration that exports nothing, so the tests
// never reach for a collector.
func baseConfig() config.TelemetryConfig {
	return config.TelemetryConfig{
		Exporter:        config.TelemetryExporterNone,
		OTLPProtocol:    config.OTLPProtocolGRPC,
		OTLPEndpoint:    "localhost:4317",
		OTLPInsecure:    true,
		Sampler:         SamplerParentBasedTraceIDRatio,
		SampleRatio:     1,
		MetricsInterval: time.Minute,
	}
}

// restoreGlobalProviders puts the global providers back to the no-op ones,
// so a test that installs the real ones does not leak into the next test.
func restoreGlobalProviders(t *testing.T) {
	t.Helper()

	t.Cleanup(func() {
		otel.SetTracerProvider(tracenoop.NewTracerProvider())
		otel.SetMeterProvider(metricnoop.NewMeterProvider())
	})
}

// resourceAttr returns the value of a resource attribute.
func resourceAttr(t *testing.T, res *resource.Resource, key string) string {
	t.Helper()

	set := attribute.NewSet(res.Attributes()...)

	value, ok := set.Value(attribute.Key(key))
	if !ok {
		t.Fatalf("expected the resource to carry %q", key)
	}
	return value.AsString()
}

func Test_Telemetry_ResourceCarriesTheProcessIdentity(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "")

	cfg := baseConfig()
	cfg.ServiceName = "requisition-system-api"
	cfg.Environment = "staging"

	telemetry := NewTelemetry("requisition-system", cfg)

	res, err := telemetry.resource(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := resourceAttr(t, res, "service.name"); got != "requisition-system-api" {
		t.Errorf("service.name = %q, want %q", got, "requisition-system-api")
	}
	if got := resourceAttr(t, res, "service.version"); got != version.Version {
		t.Errorf("service.version = %q, want %q", got, version.Version)
	}
	if got := resourceAttr(t, res, "deployment.environment.name"); got != "staging" {
		t.Errorf("deployment.environment.name = %q, want %q", got, "staging")
	}
	if got := resourceAttr(t, res, "service.instance.id"); got == "" {
		t.Error("expected the resource to carry a service instance id")
	}
}

// Test_Telemetry_ResourceKeepsTheBuiltinServiceName checks that the name
// given by the composition root is used when the environment does not
// override it.
func Test_Telemetry_ResourceKeepsTheBuiltinServiceName(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "")

	telemetry := NewTelemetry("requisition-system-worker", baseConfig())

	res, err := telemetry.resource(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := resourceAttr(t, res, "service.name"); got != "requisition-system-worker" {
		t.Errorf("service.name = %q, want %q", got, "requisition-system-worker")
	}
}

// Test_Telemetry_ResourceLetsTheEnvironmentWin checks that the standard
// OpenTelemetry environment variables override the defaults declared in code.
func Test_Telemetry_ResourceLetsTheEnvironmentWin(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "requisition-system-from-env")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "deployment.environment.name=production")

	telemetry := NewTelemetry("requisition-system", baseConfig())

	res, err := telemetry.resource(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := resourceAttr(t, res, "service.name"); got != "requisition-system-from-env" {
		t.Errorf("service.name = %q, want %q", got, "requisition-system-from-env")
	}
	if got := resourceAttr(t, res, "deployment.environment.name"); got != "production" {
		t.Errorf("deployment.environment.name = %q, want %q", got, "production")
	}
}

// Test_Telemetry_ServiceName checks the name the instrumented clients are
// told to report: the built-in one by default, the configured one when the
// environment overrides it.
func Test_Telemetry_ServiceName(t *testing.T) {
	tests := []struct {
		name          string
		configured    string
		builtin       string
		wantAsDefault string
	}{
		{
			name:          "no override",
			builtin:       "requisition-system",
			wantAsDefault: "requisition-system",
		},
		{
			name:          "configured override",
			configured:    "requisition-system-api",
			builtin:       "requisition-system",
			wantAsDefault: "requisition-system-api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := baseConfig()
			cfg.ServiceName = tt.configured

			if got := NewTelemetry(tt.builtin, cfg).ServiceName(); got != tt.wantAsDefault {
				t.Errorf("ServiceName() = %q, want %q", got, tt.wantAsDefault)
			}
		})
	}
}

func Test_Telemetry_SetupInstallsTheProviders(t *testing.T) {
	restoreGlobalProviders(t)

	before := otel.GetTracerProvider()

	telemetry := NewTelemetry("requisition-system", baseConfig())

	shutdown, err := telemetry.Setup(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() {
		if err := shutdown(context.Background()); err != nil {
			t.Errorf("shutdown: %v", err)
		}
	})

	if otel.GetTracerProvider() == before {
		t.Error("expected the tracer provider to be installed")
	}
}

// Test_Telemetry_SetupWithoutExporter checks that the process still starts
// when the telemetry is dropped.
func Test_Telemetry_SetupWithoutExporter(t *testing.T) {
	restoreGlobalProviders(t)

	cfg := baseConfig()
	cfg.Exporter = config.TelemetryExporterNone

	telemetry := NewTelemetry("requisition-system", cfg)

	shutdown, err := telemetry.Setup(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("shutdown: %v", err)
	}
}

// Test_Telemetry_SetupWithStdoutExporter checks that the exporter that was
// in use before still builds.
func Test_Telemetry_SetupWithStdoutExporter(t *testing.T) {
	restoreGlobalProviders(t)

	cfg := baseConfig()
	cfg.Exporter = config.TelemetryExporterStdout

	telemetry := NewTelemetry("requisition-system", cfg)

	shutdown, err := telemetry.Setup(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("shutdown: %v", err)
	}
}

// Test_Telemetry_SetupWithOTLPExporter checks that the OTLP exporters build
// without a collector listening: the connection is established in the
// background, so a missing collector never stops the application.
func Test_Telemetry_SetupWithOTLPExporter(t *testing.T) {
	restoreGlobalProviders(t)

	cfg := baseConfig()
	cfg.Exporter = config.TelemetryExporterOTLP

	telemetry := NewTelemetry("requisition-system", cfg)

	shutdown, err := telemetry.Setup(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The flush fails because no collector is listening, which is exactly
	// what this test checks: the exporter is built and the application
	// keeps running.
	_ = shutdown(context.Background())
}

// Test_Telemetry_SetupWithOTLPExporterOverHTTP covers the other transport.
func Test_Telemetry_SetupWithOTLPExporterOverHTTP(t *testing.T) {
	restoreGlobalProviders(t)

	cfg := baseConfig()
	cfg.Exporter = config.TelemetryExporterOTLP
	cfg.OTLPProtocol = config.OTLPProtocolHTTP

	telemetry := NewTelemetry("requisition-system", cfg)

	shutdown, err := telemetry.Setup(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = shutdown(context.Background())
}

func Test_Telemetry_SetupWithUnknownExporter(t *testing.T) {
	restoreGlobalProviders(t)

	cfg := baseConfig()
	cfg.Exporter = "carrier-pigeon"

	telemetry := NewTelemetry("requisition-system", cfg)

	if _, err := telemetry.Setup(context.Background()); err == nil {
		t.Fatal("expected an error for an unknown exporter, got none")
	}
}

func Test_Telemetry_SetupWithUnknownProtocol(t *testing.T) {
	restoreGlobalProviders(t)

	cfg := baseConfig()
	cfg.Exporter = config.TelemetryExporterOTLP
	cfg.OTLPProtocol = "smoke-signal"

	telemetry := NewTelemetry("requisition-system", cfg)

	if _, err := telemetry.Setup(context.Background()); err == nil {
		t.Fatal("expected an error for an unknown protocol, got none")
	}
}

func Test_Telemetry_SetupWithUnknownSampler(t *testing.T) {
	restoreGlobalProviders(t)

	cfg := baseConfig()
	cfg.Sampler = "sometimes"

	telemetry := NewTelemetry("requisition-system", cfg)

	if _, err := telemetry.Setup(context.Background()); err == nil {
		t.Fatal("expected an error for an unknown sampler, got none")
	}
}

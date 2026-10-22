package config

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestDBConfig_DSN(t *testing.T) {
	cfg := DBConfig{
		Host:     "db.example.com",
		Port:     "5433",
		User:     "app",
		Password: "s3cret",
		Name:     "mydb",
		SSLMode:  "require",
	}

	want := "host=db.example.com port=5433 user=app password=s3cret dbname=mydb sslmode=require"
	if got := cfg.DSN(); got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}

// unsetEnv removes a variable for the duration of the test, restoring the
// previous value (or absence) afterwards.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	old, had := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

func requireDBEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_NAME", "mydb")
}

func TestGetIntEnv_Valid(t *testing.T) {
	t.Setenv("TEST_INT_VALID", "42")

	got, err := getIntEnv("TEST_INT_VALID", 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 42 {
		t.Errorf("getIntEnv() = %d, want 42", got)
	}
}

func TestGetIntEnv_Invalid(t *testing.T) {
	t.Setenv("TEST_INT_INVALID", "not-a-number")

	_, err := getIntEnv("TEST_INT_INVALID", 7)
	if err == nil {
		t.Fatal("expected error for invalid int, got nil")
	}
	if !strings.Contains(err.Error(), "TEST_INT_INVALID") {
		t.Errorf("error should mention variable name, got: %v", err)
	}
}

func TestGetIntEnv_Default(t *testing.T) {
	unsetEnv(t, "TEST_INT_UNSET")

	got, err := getIntEnv("TEST_INT_UNSET", 9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 9 {
		t.Errorf("getIntEnv() = %d, want default 9", got)
	}
}

func TestGetTimeDurationEnv_Valid(t *testing.T) {
	t.Setenv("TEST_DURATION_VALID", "90s")

	got, err := getTimeDurationEnv("TEST_DURATION_VALID", time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 90*time.Second {
		t.Errorf("getTimeDurationEnv() = %v, want 90s", got)
	}
}

func TestGetTimeDurationEnv_Invalid(t *testing.T) {
	t.Setenv("TEST_DURATION_INVALID", "not-a-duration")

	_, err := getTimeDurationEnv("TEST_DURATION_INVALID", time.Second)
	if err == nil {
		t.Fatal("expected error for invalid duration, got nil")
	}
	if !strings.Contains(err.Error(), "TEST_DURATION_INVALID") {
		t.Errorf("error should mention variable name, got: %v", err)
	}
}

func TestGetTimeDurationEnv_Default(t *testing.T) {
	unsetEnv(t, "TEST_DURATION_UNSET")

	got, err := getTimeDurationEnv("TEST_DURATION_UNSET", 15*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 15*time.Second {
		t.Errorf("getTimeDurationEnv() = %v, want default 15s", got)
	}
}

func TestLoadDBConfig_Defaults(t *testing.T) {
	requireDBEnv(t)
	unsetEnv(t, "DB_PORT")
	unsetEnv(t, "DB_SSL_MODE")
	unsetEnv(t, "DB_MAX_OPEN_CONNECTIONS")
	unsetEnv(t, "DB_MAX_IDLE_CONNECTIONS")
	unsetEnv(t, "DB_MAX_LIFETIME_MINUTES")

	cfg, err := LoadDBConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "localhost" || cfg.User != "user" || cfg.Password != "pass" || cfg.Name != "mydb" {
		t.Errorf("expected required fields from env, got: %+v", cfg)
	}
	if cfg.Port != "5432" {
		t.Errorf("Port = %q, want default 5432", cfg.Port)
	}
	if cfg.SSLMode != "disable" {
		t.Errorf("SSLMode = %q, want default disable", cfg.SSLMode)
	}
	if cfg.MaxOpenConnections != 25 {
		t.Errorf("MaxOpenConnections = %d, want default 25", cfg.MaxOpenConnections)
	}
	if cfg.MaxIdleConnections != 10 {
		t.Errorf("MaxIdleConnections = %d, want default 10", cfg.MaxIdleConnections)
	}
	if cfg.MaxLifetimeMinutes != 30 {
		t.Errorf("MaxLifetimeMinutes = %d, want default 30", cfg.MaxLifetimeMinutes)
	}
}

func TestLoadDBConfig_ReadsValues(t *testing.T) {
	requireDBEnv(t)
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_SSL_MODE", "require")
	t.Setenv("DB_MAX_OPEN_CONNECTIONS", "50")
	t.Setenv("DB_MAX_IDLE_CONNECTIONS", "5")
	t.Setenv("DB_MAX_LIFETIME_MINUTES", "60")

	cfg, err := LoadDBConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "5433" || cfg.SSLMode != "require" {
		t.Errorf("expected port/sslmode from env, got: %+v", cfg)
	}
	if cfg.MaxOpenConnections != 50 {
		t.Errorf("MaxOpenConnections = %d, want 50", cfg.MaxOpenConnections)
	}
	if cfg.MaxIdleConnections != 5 {
		t.Errorf("MaxIdleConnections = %d, want 5", cfg.MaxIdleConnections)
	}
	if cfg.MaxLifetimeMinutes != 60 {
		t.Errorf("MaxLifetimeMinutes = %d, want 60", cfg.MaxLifetimeMinutes)
	}
}

func TestLoadDBConfig_InvalidInt(t *testing.T) {
	requireDBEnv(t)
	t.Setenv("DB_MAX_OPEN_CONNECTIONS", "abc")

	_, err := LoadDBConfig()
	if err == nil {
		t.Fatal("expected error for invalid int env, got nil")
	}
	if !strings.Contains(err.Error(), "DB_MAX_OPEN_CONNECTIONS") {
		t.Errorf("error should mention DB_MAX_OPEN_CONNECTIONS, got: %v", err)
	}
}

func TestLoadServerConfig_Defaults(t *testing.T) {
	unsetEnv(t, "SERVER_PORT")
	unsetEnv(t, "SERVER_READ_TIMEOUT")
	unsetEnv(t, "SERVER_WRITE_TIMEOUT")
	unsetEnv(t, "SERVER_IDLE_TIMEOUT")

	cfg, err := LoadServerConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Address != ":8080" {
		t.Errorf("Address = %q, want default :8080", cfg.Address)
	}
	if cfg.ReadTimeout != 5*time.Second {
		t.Errorf("ReadTimeout = %v, want 5s", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 10*time.Second {
		t.Errorf("WriteTimeout = %v, want 10s", cfg.WriteTimeout)
	}
	if cfg.IdleTimeout != 120*time.Second {
		t.Errorf("IdleTimeout = %v, want 120s", cfg.IdleTimeout)
	}
}

func TestLoadServerConfig_InvalidDuration(t *testing.T) {
	t.Setenv("SERVER_READ_TIMEOUT", "not-a-duration")

	_, err := LoadServerConfig()
	if err == nil {
		t.Fatal("expected error for invalid duration env, got nil")
	}
	if !strings.Contains(err.Error(), "SERVER_READ_TIMEOUT") {
		t.Errorf("error should mention SERVER_READ_TIMEOUT, got: %v", err)
	}
}

func requireJWTEnv(t *testing.T) string {
	t.Helper()
	secret := strings.Repeat("a", minJWTSecretLength)
	t.Setenv("JWT_SECRET", secret)
	return secret
}

func TestLoadJWTConfig_Defaults(t *testing.T) {
	secret := requireJWTEnv(t)
	unsetEnv(t, "JWT_ISSUER")
	unsetEnv(t, "JWT_AUDIENCE")
	unsetEnv(t, "JWT_ACCESS_TOKEN_TTL")
	unsetEnv(t, "JWT_REFRESH_TOKEN_TTL")
	unsetEnv(t, "JWT_LEEWAY")

	cfg, err := LoadJWTConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Secret != secret {
		t.Errorf("Secret = %q, want %q", cfg.Secret, secret)
	}
	if cfg.Issuer != "requisition-system" {
		t.Errorf("Issuer = %q, want default requisition-system", cfg.Issuer)
	}
	if cfg.Audience != "requisition-system" {
		t.Errorf("Audience = %q, want default requisition-system", cfg.Audience)
	}
	if cfg.AccessTokenTTL != 15*time.Minute {
		t.Errorf("AccessTokenTTL = %v, want 15m", cfg.AccessTokenTTL)
	}
	if cfg.RefreshTokenTTL != 7*24*time.Hour {
		t.Errorf("RefreshTokenTTL = %v, want 168h", cfg.RefreshTokenTTL)
	}
	if cfg.Leeway != 30*time.Second {
		t.Errorf("Leeway = %v, want 30s", cfg.Leeway)
	}
}

func TestLoadJWTConfig_ReadsValues(t *testing.T) {
	requireJWTEnv(t)
	t.Setenv("JWT_ISSUER", "my-issuer")
	t.Setenv("JWT_AUDIENCE", "my-audience")
	t.Setenv("JWT_ACCESS_TOKEN_TTL", "5m")
	t.Setenv("JWT_REFRESH_TOKEN_TTL", "24h")
	t.Setenv("JWT_LEEWAY", "10s")

	cfg, err := LoadJWTConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Issuer != "my-issuer" || cfg.Audience != "my-audience" {
		t.Errorf("expected issuer/audience from env, got: %+v", cfg)
	}
	if cfg.AccessTokenTTL != 5*time.Minute || cfg.RefreshTokenTTL != 24*time.Hour {
		t.Errorf("expected TTLs from env, got: %+v", cfg)
	}
	if cfg.Leeway != 10*time.Second {
		t.Errorf("Leeway = %v, want 10s", cfg.Leeway)
	}
}

func TestLoadJWTConfig_SecretTooShort(t *testing.T) {
	t.Setenv("JWT_SECRET", "too-short")

	_, err := LoadJWTConfig()
	if !errors.Is(err, ErrJWTSecretTooShort) {
		t.Fatalf("expected ErrJWTSecretTooShort, got: %v", err)
	}
}

func TestLoadRedisConfig_Defaults(t *testing.T) {
	t.Setenv("REDIS_HOST", "localhost")
	t.Setenv("REDIS_PORT", "6379")
	unsetEnv(t, "REDIS_USERNAME")
	unsetEnv(t, "REDIS_PASSWORD")
	unsetEnv(t, "REDIS_DB")

	cfg, err := LoadRedisConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "localhost" || cfg.Port != "6379" {
		t.Errorf("expected host/port from env, got: %+v", cfg)
	}
	if cfg.Username != "" || cfg.Password != "" {
		t.Errorf("expected empty credentials, got: %+v", cfg)
	}
	if cfg.DB != 0 {
		t.Errorf("DB = %d, want 0", cfg.DB)
	}
}

func TestLoadRedisConfig_ReadsValues(t *testing.T) {
	t.Setenv("REDIS_HOST", "redis.internal")
	t.Setenv("REDIS_PORT", "6380")
	t.Setenv("REDIS_USERNAME", "app")
	t.Setenv("REDIS_PASSWORD", "s3cret")
	t.Setenv("REDIS_DB", "2")

	cfg, err := LoadRedisConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "redis.internal" || cfg.Port != "6380" ||
		cfg.Username != "app" || cfg.Password != "s3cret" {
		t.Errorf("unexpected config: %+v", cfg)
	}
	if cfg.DB != 2 {
		t.Errorf("DB = %d, want 2", cfg.DB)
	}
}

func TestLoadRedisConfig_MissingHost(t *testing.T) {
	unsetEnv(t, "REDIS_HOST")
	t.Setenv("REDIS_PORT", "6379")

	if _, err := LoadRedisConfig(); err == nil {
		t.Fatal("expected error when REDIS_HOST is missing")
	}
}

func TestLoadRedisConfig_InvalidDB(t *testing.T) {
	t.Setenv("REDIS_HOST", "localhost")
	t.Setenv("REDIS_PORT", "6379")
	t.Setenv("REDIS_DB", "not-a-number")

	if _, err := LoadRedisConfig(); err == nil {
		t.Fatal("expected error for non-numeric REDIS_DB")
	}
}

func TestLoadJWTConfig_RefreshTTLNotLongerThanAccess(t *testing.T) {
	requireJWTEnv(t)
	t.Setenv("JWT_ACCESS_TOKEN_TTL", "1h")
	t.Setenv("JWT_REFRESH_TOKEN_TTL", "5m")

	_, err := LoadJWTConfig()
	if !errors.Is(err, ErrJWTRefreshTTLNotLongerThanAccess) {
		t.Fatalf("expected ErrJWTRefreshTTLNotLongerThanAccess, got: %v", err)
	}
}

func requireEmailEnv(t *testing.T) {
	t.Helper()
	t.Setenv("RESEND_API_KEY", "re_test_key")
}

func TestLoadEmailConfig_Defaults(t *testing.T) {
	requireEmailEnv(t)
	unsetEnv(t, "EMAIL_FROM")
	unsetEnv(t, "EMAIL_FROM_NAME")

	cfg, err := LoadEmailConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.APIKey != "re_test_key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "re_test_key")
	}
	if cfg.From != "onboarding@resend.dev" {
		t.Errorf("From = %q, want default onboarding@resend.dev", cfg.From)
	}
	if cfg.FromName != "Requisition System" {
		t.Errorf("FromName = %q, want default Requisition System", cfg.FromName)
	}
}

func TestLoadEmailConfig_ReadsValues(t *testing.T) {
	requireEmailEnv(t)
	t.Setenv("EMAIL_FROM", "orders@example.com")
	t.Setenv("EMAIL_FROM_NAME", "Orders")

	cfg, err := LoadEmailConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.From != "orders@example.com" || cfg.FromName != "Orders" {
		t.Errorf("expected sender from env, got: %+v", cfg)
	}
}

func TestLoadEmailConfig_MissingAPIKey(t *testing.T) {
	unsetEnv(t, "RESEND_API_KEY")

	if _, err := LoadEmailConfig(); err == nil {
		t.Fatal("expected error when RESEND_API_KEY is missing")
	}
}

// requireTelemetryEnv clears every telemetry variable, so a test starts from
// a known state whatever the environment of the developer holds.
func requireTelemetryEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"TELEMETRY_EXPORTER",
		"OTEL_SERVICE_NAME",
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_PROTOCOL",
		"OTEL_EXPORTER_OTLP_INSECURE",
		"OTEL_TRACES_SAMPLER",
		"OTEL_TRACES_SAMPLER_ARG",
		"OTEL_METRIC_EXPORT_INTERVAL",
		"APP_ENV",
	} {
		unsetEnv(t, key)
	}
}

func TestLoadTelemetryConfig_Defaults(t *testing.T) {
	requireTelemetryEnv(t)

	cfg, err := LoadTelemetryConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Exporter != TelemetryExporterStdout {
		t.Errorf("Exporter = %q, want %q", cfg.Exporter, TelemetryExporterStdout)
	}
	if cfg.ServiceName != "" {
		t.Errorf("ServiceName = %q, want an empty default", cfg.ServiceName)
	}
	if cfg.Environment != "development" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "development")
	}
	if cfg.OTLPEndpoint != "localhost:4317" {
		t.Errorf("OTLPEndpoint = %q, want %q", cfg.OTLPEndpoint, "localhost:4317")
	}
	if cfg.OTLPProtocol != OTLPProtocolGRPC {
		t.Errorf("OTLPProtocol = %q, want %q", cfg.OTLPProtocol, OTLPProtocolGRPC)
	}
	if !cfg.OTLPInsecure {
		t.Error("OTLPInsecure = false, want true")
	}
	if cfg.Sampler != "parentbased_traceidratio" {
		t.Errorf("Sampler = %q, want %q", cfg.Sampler, "parentbased_traceidratio")
	}
	if cfg.SampleRatio != 1 {
		t.Errorf("SampleRatio = %v, want 1", cfg.SampleRatio)
	}
	if cfg.MetricsInterval != 60*time.Second {
		t.Errorf("MetricsInterval = %v, want %v", cfg.MetricsInterval, 60*time.Second)
	}
}

func TestLoadTelemetryConfig_ReadsValues(t *testing.T) {
	requireTelemetryEnv(t)
	t.Setenv("TELEMETRY_EXPORTER", TelemetryExporterOTLP)
	t.Setenv("OTEL_SERVICE_NAME", "requisition-system-api")
	t.Setenv("APP_ENV", "production")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "collector:4317")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", OTLPProtocolHTTP)
	t.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "false")
	t.Setenv("OTEL_TRACES_SAMPLER", "traceidratio")
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "0.25")
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "15000")

	cfg, err := LoadTelemetryConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Exporter != TelemetryExporterOTLP {
		t.Errorf("Exporter = %q, want %q", cfg.Exporter, TelemetryExporterOTLP)
	}
	if cfg.ServiceName != "requisition-system-api" {
		t.Errorf("ServiceName = %q, want %q", cfg.ServiceName, "requisition-system-api")
	}
	if cfg.Environment != "production" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "production")
	}
	if cfg.OTLPEndpoint != "collector:4317" {
		t.Errorf("OTLPEndpoint = %q, want %q", cfg.OTLPEndpoint, "collector:4317")
	}
	if cfg.OTLPProtocol != OTLPProtocolHTTP {
		t.Errorf("OTLPProtocol = %q, want %q", cfg.OTLPProtocol, OTLPProtocolHTTP)
	}
	if cfg.OTLPInsecure {
		t.Error("OTLPInsecure = true, want false")
	}
	if cfg.Sampler != "traceidratio" {
		t.Errorf("Sampler = %q, want %q", cfg.Sampler, "traceidratio")
	}
	if cfg.SampleRatio != 0.25 {
		t.Errorf("SampleRatio = %v, want 0.25", cfg.SampleRatio)
	}
	if cfg.MetricsInterval != 15*time.Second {
		t.Errorf("MetricsInterval = %v, want %v", cfg.MetricsInterval, 15*time.Second)
	}
}

func TestLoadTelemetryConfig_UnsupportedExporter(t *testing.T) {
	requireTelemetryEnv(t)
	t.Setenv("TELEMETRY_EXPORTER", "carrier-pigeon")

	_, err := LoadTelemetryConfig()
	if err == nil {
		t.Fatal("expected an error for an unknown exporter, got none")
	}
	if !errors.Is(err, ErrTelemetryExporterUnsupported) {
		t.Errorf("expected ErrTelemetryExporterUnsupported, got: %v", err)
	}
}

func TestLoadTelemetryConfig_UnsupportedProtocol(t *testing.T) {
	requireTelemetryEnv(t)
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "smoke-signal")

	_, err := LoadTelemetryConfig()
	if err == nil {
		t.Fatal("expected an error for an unknown protocol, got none")
	}
	if !errors.Is(err, ErrTelemetryProtocolUnsupported) {
		t.Errorf("expected ErrTelemetryProtocolUnsupported, got: %v", err)
	}
}

func TestLoadTelemetryConfig_SampleRatioOutOfRange(t *testing.T) {
	requireTelemetryEnv(t)
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "1.5")

	_, err := LoadTelemetryConfig()
	if err == nil {
		t.Fatal("expected an error for a sample ratio above one, got none")
	}
	if !errors.Is(err, ErrTelemetrySampleRatioOutOfRange) {
		t.Errorf("expected ErrTelemetrySampleRatioOutOfRange, got: %v", err)
	}
}

func TestLoadTelemetryConfig_InvalidSampleRatio(t *testing.T) {
	requireTelemetryEnv(t)
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "half")

	if _, err := LoadTelemetryConfig(); err == nil {
		t.Fatal("expected an error for an invalid sample ratio, got none")
	}
}

func TestLoadTelemetryConfig_InvalidBool(t *testing.T) {
	requireTelemetryEnv(t)
	t.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "maybe")

	if _, err := LoadTelemetryConfig(); err == nil {
		t.Fatal("expected an error for an invalid boolean, got none")
	}
}

func TestLoadTelemetryConfig_InvalidInterval(t *testing.T) {
	requireTelemetryEnv(t)
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "soon")

	if _, err := LoadTelemetryConfig(); err == nil {
		t.Fatal("expected an error for an invalid interval, got none")
	}
}

// The interval is counted in milliseconds, as the OpenTelemetry specification
// defines it. A duration is not a valid value, and the SDK that reads the
// same variable rejects it too.
func TestLoadTelemetryConfig_IntervalIsNotADuration(t *testing.T) {
	requireTelemetryEnv(t)
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "15s")

	if _, err := LoadTelemetryConfig(); err == nil {
		t.Fatal("expected an error for an interval written as a duration, got none")
	}
}

func TestLoadTelemetryConfig_NonPositiveInterval(t *testing.T) {
	requireTelemetryEnv(t)
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "0")

	_, err := LoadTelemetryConfig()
	if err == nil {
		t.Fatal("expected an error for a non positive interval, got none")
	}
	if !errors.Is(err, ErrTelemetryMetricsIntervalInvalid) {
		t.Errorf("expected ErrTelemetryMetricsIntervalInvalid, got: %v", err)
	}
}

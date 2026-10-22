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

type Config struct {
	App       AppConfig
	API       APIConfig
	DB        DBConfig
	Redis     RedisConfig
	Server    ServerConfig
	JWT       JWTConfig
	Email     EmailConfig
	Frontend  FrontendConfig
	Worker    WorkerConfig
	Telemetry TelemetryConfig
}

type AppConfig struct {
	Env string // "development", "staging", "production"
}

type APIConfig struct {
	BaseURL string // e.g. "/api/v1" or "http://localhost:8080"
}

type DBConfig struct {
	Host               string
	Port               string
	User               string
	Password           string
	Name               string
	SSLMode            string
	MaxOpenConnections int
	MaxIdleConnections int
	MaxLifetimeMinutes int
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

type RedisConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DB       int
}

type ServerConfig struct {
	Address      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type JWTConfig struct {
	Secret          string
	Issuer          string
	Audience        string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration // also the Redis refresh-session TTL
	Leeway          time.Duration
}

type EmailConfig struct {
	APIKey   string // Resend API key
	From     string // verified sender address, e.g. "onboarding@resend.dev"
	FromName string
}

type FrontendConfig struct {
	BaseURL string
}

// WorkerConfig configures the asynq worker process and the tasks enqueued
// for it.
type WorkerConfig struct {
	Concurrency     int           // parallel task handlers
	Queue           string        // queue the worker drains and tasks are pushed to
	TaskMaxRetry    int           // delivery attempts before a task is given up on
	TaskTimeout     time.Duration // per-attempt processing timeout
	ShutdownTimeout time.Duration // grace period for in-flight tasks on shutdown
}

// TelemetryConfig configures how the OpenTelemetry traces and metrics leave
// the process. The variable names follow the OpenTelemetry specification, so
// the application can be pointed at any OTLP backend without a code change.
type TelemetryConfig struct {
	ServiceName     string        // OTEL_SERVICE_NAME; empty keeps the built-in name
	Environment     string        // deployment.environment.name of the resource
	Exporter        string        // stdout, otlp or none
	OTLPEndpoint    string        // OTEL_EXPORTER_OTLP_ENDPOINT
	OTLPProtocol    string        // OTEL_EXPORTER_OTLP_PROTOCOL: grpc or http
	OTLPInsecure    bool          // OTEL_EXPORTER_OTLP_INSECURE
	Sampler         string        // OTEL_TRACES_SAMPLER
	SampleRatio     float64       // OTEL_TRACES_SAMPLER_ARG
	MetricsInterval time.Duration // OTEL_METRIC_EXPORT_INTERVAL
}

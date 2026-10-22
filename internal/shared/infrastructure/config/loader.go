package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/joho/godotenv"
)

const minJWTSecretLength = 32

var (
	ErrJWTSecretTooShort                = errors.New("JWT secret must be at least 32 characters long")
	ErrJWTRefreshTTLNotLongerThanAccess = errors.New(
		"JWT refresh token TTL must be greater than the access token TTL",
	)
	ErrTelemetryExporterUnsupported = errors.New(
		"telemetry exporter must be one of: stdout, otlp, none",
	)
	ErrTelemetryProtocolUnsupported = errors.New(
		"telemetry OTLP protocol must be one of: grpc, http",
	)
	ErrTelemetrySamplerUnsupported = errors.New(
		"telemetry sampler must be one of: always_on, always_off, " +
			"parentbased_always_on, parentbased_traceidratio, traceidratio",
	)
	ErrTelemetrySampleRatioOutOfRange = errors.New(
		"telemetry sample ratio must be between 0 and 1",
	)
	ErrTelemetryMetricsIntervalInvalid = errors.New(
		"telemetry metric export interval must be positive",
	)
)

func Load() (*Config, error) {
	_ = godotenv.Load()
	appConfig, err := LoadAppConfig()
	if err != nil {
		return nil, err
	}

	dbConfig, err := LoadDBConfig()
	if err != nil {
		return nil, err
	}

	apiConfig, err := LoadAPIConfig()
	if err != nil {
		return nil, err
	}
	serverConfig, err := LoadServerConfig()
	if err != nil {
		return nil, err
	}

	jwtConfig, err := LoadJWTConfig()
	if err != nil {
		return nil, err
	}

	redisConfig, err := LoadRedisConfig()
	if err != nil {
		return nil, err
	}

	emailConfig, err := LoadEmailConfig()
	if err != nil {
		return nil, err
	}

	frontendConfig, err := LoadFrontendConfig()
	if err != nil {
		return nil, err
	}

	workerConfig, err := LoadWorkerConfig()
	if err != nil {
		return nil, err
	}

	telemetryConfig, err := LoadTelemetryConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		App:       *appConfig,
		API:       *apiConfig,
		DB:        *dbConfig,
		Redis:     *redisConfig,
		Server:    *serverConfig,
		JWT:       *jwtConfig,
		Email:     *emailConfig,
		Frontend:  *frontendConfig,
		Worker:    *workerConfig,
		Telemetry: *telemetryConfig,
	}, nil
}

func LoadServerConfig() (*ServerConfig, error) {
	address := getEnv("SERVER_PORT", ":8080")

	readTimeout, err := getTimeDurationEnv("SERVER_READ_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}
	writeTimeout, err := getTimeDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, err
	}
	idleTimeout, err := getTimeDurationEnv("SERVER_IDLE_TIMEOUT", 120*time.Second)
	if err != nil {
		return nil, err
	}

	return &ServerConfig{
		Address:      address,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}, nil
}

func LoadDBConfig() (*DBConfig, error) {
	_ = godotenv.Load()
	dbHost, err := requireEnv("DB_HOST")
	if err != nil {
		return nil, err
	}

	dbUser, err := requireEnv("DB_USER")
	if err != nil {
		return nil, err
	}

	dbPassword, err := requireEnv("DB_PASSWORD")
	if err != nil {
		return nil, err
	}

	dbName, err := requireEnv("DB_NAME")
	if err != nil {
		return nil, err
	}

	maxOpenConnections, err := getIntEnv("DB_MAX_OPEN_CONNECTIONS", 25)
	if err != nil {
		return nil, err
	}
	maxIdleConnections, err := getIntEnv("DB_MAX_IDLE_CONNECTIONS", 10)
	if err != nil {
		return nil, err
	}
	maxLifetimeMinutes, err := getIntEnv("DB_MAX_LIFETIME_MINUTES", 30)
	if err != nil {
		return nil, err
	}

	return &DBConfig{
		Host:               dbHost,
		Port:               getEnv("DB_PORT", "5432"),
		User:               dbUser,
		Password:           dbPassword,
		Name:               dbName,
		SSLMode:            getEnv("DB_SSL_MODE", "disable"),
		MaxOpenConnections: maxOpenConnections,
		MaxIdleConnections: maxIdleConnections,
		MaxLifetimeMinutes: maxLifetimeMinutes,
	}, nil
}

func LoadAppConfig() (*AppConfig, error) {
	env := getEnv("APP_ENV", "development")

	return &AppConfig{
		Env: env,
	}, nil
}

func LoadAPIConfig() (*APIConfig, error) {
	baseURL := getEnv("API_BASE_URL", "/api/v1")

	return &APIConfig{
		BaseURL: baseURL,
	}, nil
}

func LoadJWTConfig() (*JWTConfig, error) {
	secret, err := requireEnv("JWT_SECRET")
	if err != nil {
		return nil, err
	}

	if len(secret) < minJWTSecretLength {
		return nil, ErrJWTSecretTooShort
	}

	accessTokenTTL, err := getTimeDurationEnv("JWT_ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return nil, err
	}

	refreshTokenTTL, err := getTimeDurationEnv("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	if refreshTokenTTL <= accessTokenTTL {
		return nil, ErrJWTRefreshTTLNotLongerThanAccess
	}

	leeway, err := getTimeDurationEnv("JWT_LEEWAY", 30*time.Second)
	if err != nil {
		return nil, err
	}

	return &JWTConfig{
		Secret:          secret,
		Issuer:          getEnv("JWT_ISSUER", "requisition-system"),
		Audience:        getEnv("JWT_AUDIENCE", "requisition-system"),
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
		Leeway:          leeway,
	}, nil
}

func LoadRedisConfig() (*RedisConfig, error) {
	host, err := requireEnv("REDIS_HOST")
	if err != nil {
		return nil, err
	}

	db, err := getIntEnv("REDIS_DB", 0)
	if err != nil {
		return nil, err
	}
	port, err := requireEnv("REDIS_PORT")
	if err != nil {
		return nil, err
	}

	return &RedisConfig{
		Host:     host,
		Port:     port,
		Username: getEnv("REDIS_USERNAME", ""),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       db,
	}, nil
}

// LoadEmailConfig loads the Resend (https://resend.com) credentials. The API
// key is required: invitations cannot be delivered without it.
func LoadEmailConfig() (*EmailConfig, error) {
	apiKey, err := requireEnv("RESEND_API_KEY")
	if err != nil {
		return nil, err
	}

	return &EmailConfig{
		APIKey:   apiKey,
		From:     getEnv("EMAIL_FROM", "onboarding@resend.dev"),
		FromName: getEnv("EMAIL_FROM_NAME", "Requisition System"),
	}, nil
}

func LoadFrontendConfig() (*FrontendConfig, error) {
	return &FrontendConfig{
		BaseURL: getEnv("FRONTEND_BASE_URL", "http://localhost:3000"),
	}, nil
}

// LoadWorkerConfig loads the background worker settings. Defaults keep a
// single-queue setup running out of the box.
func LoadWorkerConfig() (*WorkerConfig, error) {
	concurrency, err := getIntEnv("WORKER_CONCURRENCY", 10)
	if err != nil {
		return nil, err
	}

	taskMaxRetry, err := getIntEnv("WORKER_TASK_MAX_RETRY", 5)
	if err != nil {
		return nil, err
	}

	taskTimeout, err := getTimeDurationEnv("WORKER_TASK_TIMEOUT", 30*time.Second)
	if err != nil {
		return nil, err
	}

	shutdownTimeout, err := getTimeDurationEnv("WORKER_SHUTDOWN_TIMEOUT", 8*time.Second)
	if err != nil {
		return nil, err
	}

	return &WorkerConfig{
		Concurrency:     concurrency,
		Queue:           getEnv("WORKER_QUEUE", "default"),
		TaskMaxRetry:    taskMaxRetry,
		TaskTimeout:     taskTimeout,
		ShutdownTimeout: shutdownTimeout,
	}, nil
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

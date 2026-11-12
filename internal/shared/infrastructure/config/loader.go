package config

import (
	"errors"
	"fmt"

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
	ErrProductionPostgresTLSRequired = errors.New(
		"production requires PostgreSQL DB_SSL_MODE=verify-full",
	)
	ErrProductionRedisTLSRequired = errors.New(
		"production requires REDIS_TLS=true",
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

	var jwtConfig *JWTConfig
	if appConfig.Env == ProdEnv {
		jwtConfig, err = ProdLoadJWTConfig()
	} else {
		jwtConfig, err = LoadJWTConfig()
	}
	if err != nil {
		return nil, err
	}

	redisConfig, err := LoadRedisConfig()
	if err != nil {
		return nil, err
	}
	if err := validateProductionTLS(
		appConfig.Env,
		dbConfig.SSLMode,
		redisConfig.TLS,
	); err != nil {
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

func validateProductionTLS(env Environment, dbSSLMode string, redisTLS bool) error {
	if env != ProdEnv {
		return nil
	}
	if dbSSLMode != "verify-full" {
		return fmt.Errorf("%w: got %q", ErrProductionPostgresTLSRequired, dbSSLMode)
	}
	if !redisTLS {
		return ErrProductionRedisTLSRequired
	}
	return nil
}

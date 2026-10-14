package config

import (
	"errors"
	"time"

	"github.com/joho/godotenv"
)

const minJWTSecretLength = 32

var (
	ErrJWTSecretTooShort                = errors.New("JWT secret must be at least 32 characters long")
	ErrJWTRefreshTTLNotLongerThanAccess = errors.New(
		"JWT refresh token TTL must be greater than the access token TTL",
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

	return &Config{
		App:      *appConfig,
		API:      *apiConfig,
		DB:       *dbConfig,
		Redis:    *redisConfig,
		Server:   *serverConfig,
		JWT:      *jwtConfig,
		Email:    *emailConfig,
		Frontend: *frontendConfig,
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

func LoadEmailConfig() (*EmailConfig, error) {
	host := getEnv("EMAIL_HOST", "localhost")
	port, err := getIntEnv("EMAIL_PORT", 1025)
	if err != nil {
		return nil, err
	}
	return &EmailConfig{
		Host:     host,
		Port:     port,
		Username: getEnv("EMAIL_USERNAME", ""),
		Password: getEnv("EMAIL_PASSWORD", ""),
		From:     getEnv("EMAIL_FROM", "noreply@requisition-system.local"),
		FromName: getEnv("EMAIL_FROM_NAME", "Requisition System"),
		SSL:      getEnv("EMAIL_SSL", "false") == "true",
	}, nil
}

func LoadFrontendConfig() (*FrontendConfig, error) {
	return &FrontendConfig{
		BaseURL: getEnv("FRONTEND_BASE_URL", "http://localhost:3000"),
	}, nil
}

package config

import (
	"time"

	"github.com/joho/godotenv"
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

	return &Config{
		App:    *appConfig,
		API:    *apiConfig,
		DB:     *dbConfig,
		Server: *serverConfig,
	}, nil
}

func LoadServerConfig() (*ServerConfig, error) {
	address := getEnv("SERVER_PORT", ":8080")
	readTimeout := getTimeDurationEnv("SERVER_READ_TIMEOUT", 5*time.Second)
	writeTimeout := getTimeDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second)
	idleTimeout := getTimeDurationEnv("SERVER_IDLE_TIMEOUT", 120*time.Second)

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

	return &DBConfig{
		Host:     dbHost,
		Port:     getEnv("DB_PORT", "5432"),
		User:     dbUser,
		Password: dbPassword,
		Name:     dbName,
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
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

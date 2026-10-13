package config

import "time"

type Config struct {
	App    AppConfig
	API    APIConfig
	DB     DBConfig
	Redis  RedisConfig
	Server ServerConfig
	JWT    JWTConfig
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

type RedisConfig struct {
	Host  string
	Port  string
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

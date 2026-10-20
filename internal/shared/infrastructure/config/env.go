package config

import "time"

type Config struct {
	App      AppConfig
	API      APIConfig
	DB       DBConfig
	Redis    RedisConfig
	Server   ServerConfig
	JWT      JWTConfig
	Email    EmailConfig
	Frontend FrontendConfig
	Worker   WorkerConfig
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

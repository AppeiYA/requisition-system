package config

import "time"

type Config struct {
	App    AppConfig
	API    APIConfig
	DB     DBConfig
	Server ServerConfig
}

type AppConfig struct {
	Env string // "development", "staging", "production"
}

type APIConfig struct {
	BaseURL string // e.g. "/api/v1" or "http://localhost:8080"
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type ServerConfig struct {
	Address      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

package config

import "time"

type ServerConfig struct {
	Address      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
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
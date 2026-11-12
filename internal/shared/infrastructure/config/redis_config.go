package config

import "fmt"

type RedisConfig struct {
	Host          string
	Port          string
	Username      string
	Password      string
	DB            int
	TLS           bool
	TLSCACert     string
	TLSServerName string
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
	tls, err := getBoolEnv("REDIS_TLS", false)
	if err != nil {
		return nil, err
	}

	caCert := getEnv("REDIS_CA_CERT", "")
	if caCert != "" {
		if err := validateCACert(caCert); err != nil {
			return nil, fmt.Errorf("REDIS_CA_CERT: %w", err)
		}
	}

	return &RedisConfig{
		Host:          host,
		Port:          port,
		Username:      getEnv("REDIS_USERNAME", ""),
		Password:      getEnv("REDIS_PASSWORD", ""),
		DB:            db,
		TLS:           tls,
		TLSCACert:     caCert,
		TLSServerName: getEnv("REDIS_TLS_SERVER_NAME", host),
	}, nil
}

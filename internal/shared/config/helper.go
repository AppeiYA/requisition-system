package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func returnEnvNotFoundError(variable string) error {
	return fmt.Errorf("environment variable %s is required but not set", variable)
}

func returnEnvInvalidError(variable, expectedType string) error {
	return fmt.Errorf("environment variable %s is set but is not a valid %s", variable, expectedType)
}

func getEnv(variable, defaultValue string) string {
	if val, ok := os.LookupEnv(variable); ok {
		return val
	} else {
		return defaultValue
	}
}

func requireEnv(variable string) (string, error) {
	if val, ok := os.LookupEnv(variable); ok {
		return val, nil
	} else {
		return "", returnEnvNotFoundError(variable)
	}
}

func getIntEnv(variable string, defaultValue int) int {
	if val, ok := os.LookupEnv(variable); ok {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getTimeDurationEnv(variable string, defaultValue time.Duration) time.Duration {
	if val, ok := os.LookupEnv(variable); ok {
		if durationVal, err := time.ParseDuration(val); err == nil {
			return durationVal
		}
	}
	return defaultValue
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

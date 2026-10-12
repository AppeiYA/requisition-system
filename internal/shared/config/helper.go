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

func getIntEnv(variable string, defaultValue int) (int, error) {
	val, ok := os.LookupEnv(variable)
	if !ok {
		return defaultValue, nil
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return 0, returnEnvInvalidError(variable, "int")
	}
	return intVal, nil
}

func getTimeDurationEnv(variable string, defaultValue time.Duration) (time.Duration, error) {
	val, ok := os.LookupEnv(variable)
	if !ok {
		return defaultValue, nil
	}
	durationVal, err := time.ParseDuration(val)
	if err != nil {
		return 0, returnEnvInvalidError(variable, "duration")
	}
	return durationVal, nil
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

package config

import (
	"fmt"
	"math"
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

// getDurationMillisEnv reads a duration written in milliseconds.
//
// The intervals the OpenTelemetry specification defines are counted in
// milliseconds, and the SDK reads the very same variables the application
// does: a value written as a duration ("10s") is rejected by the SDK, and a
// value written as the specification asks ("10000") has to be accepted here
// as well. The unit is therefore the one the specification defines.
func getDurationMillisEnv(variable string, defaultValue time.Duration) (time.Duration, error) {
	val, ok := os.LookupEnv(variable)
	if !ok {
		return defaultValue, nil
	}
	millisVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, returnEnvInvalidError(variable, "number of milliseconds")
	}
	if millisVal > int64(math.MaxInt64/time.Millisecond) {
		return 0, returnEnvInvalidError(variable, "number of milliseconds")
	}
	return time.Duration(millisVal) * time.Millisecond, nil
}

func getFloatEnv(variable string, defaultValue float64) (float64, error) {
	val, ok := os.LookupEnv(variable)
	if !ok {
		return defaultValue, nil
	}
	floatVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, returnEnvInvalidError(variable, "float")
	}
	return floatVal, nil
}

func getBoolEnv(variable string, defaultValue bool) (bool, error) {
	val, ok := os.LookupEnv(variable)
	if !ok {
		return defaultValue, nil
	}
	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		return false, returnEnvInvalidError(variable, "bool")
	}
	return boolVal, nil
}

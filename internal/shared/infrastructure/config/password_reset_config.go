package config

import (
	"fmt"
	"time"
)

// PasswordResetConfig controls the self-service password reset flow.
type PasswordResetConfig struct {
	// Enabled gates the forgot/reset routes and the email producer. It exists
	// so the backend can be deployed before the frontend reset page exists.
	Enabled bool

	// TokenTTL is how long a reset link stays valid.
	TokenTTL time.Duration

	// RequestCooldown is the minimum gap between two reset emails for the same
	// account.
	RequestCooldown time.Duration

	// MaxPerHour caps reset requests per account in a rolling hour.
	MaxPerHour int

	// CleanupRetention is how long terminal token rows are kept before the
	// cleanup task deletes them.
	CleanupRetention time.Duration
}

func LoadPasswordResetConfig() (*PasswordResetConfig, error) {
	enabled, err := getBoolEnv("PASSWORD_RESET_ENABLED", true)
	if err != nil {
		return nil, err
	}

	tokenTTL, err := getTimeDurationEnv("PASSWORD_RESET_TOKEN_TTL", 30*time.Minute)
	if err != nil {
		return nil, err
	}
	if tokenTTL <= 0 {
		return nil, fmt.Errorf("PASSWORD_RESET_TOKEN_TTL must be positive")
	}

	requestCooldown, err := getTimeDurationEnv("PASSWORD_RESET_REQUEST_COOLDOWN", time.Minute)
	if err != nil {
		return nil, err
	}
	if requestCooldown <= 0 {
		return nil, fmt.Errorf("PASSWORD_RESET_REQUEST_COOLDOWN must be positive")
	}

	maxPerHour, err := getIntEnv("PASSWORD_RESET_MAX_PER_HOUR", 5)
	if err != nil {
		return nil, err
	}
	if maxPerHour <= 0 {
		return nil, fmt.Errorf("PASSWORD_RESET_MAX_PER_HOUR must be positive")
	}

	cleanupRetention, err := getTimeDurationEnv("PASSWORD_RESET_CLEANUP_RETENTION", 7*24*time.Hour)
	if err != nil {
		return nil, err
	}
	if cleanupRetention <= 0 {
		return nil, fmt.Errorf("PASSWORD_RESET_CLEANUP_RETENTION must be positive")
	}
	if cleanupRetention < tokenTTL {
		return nil, fmt.Errorf("PASSWORD_RESET_CLEANUP_RETENTION must be at least PASSWORD_RESET_TOKEN_TTL")
	}

	return &PasswordResetConfig{
		Enabled:          enabled,
		TokenTTL:         tokenTTL,
		RequestCooldown:  requestCooldown,
		MaxPerHour:       maxPerHour,
		CleanupRetention: cleanupRetention,
	}, nil
}

package config

import "fmt"

type FrontendConfig struct {
	BaseURL           string
	PasswordResetPath string
}

func LoadFrontendConfig() (*FrontendConfig, error) {
	passwordResetPath := getEnv("FRONTEND_PASSWORD_RESET_PATH", "/reset-password")
	if len(passwordResetPath) == 0 || passwordResetPath[0] != '/' {
		return nil, fmt.Errorf("FRONTEND_PASSWORD_RESET_PATH must start with '/'")
	}

	return &FrontendConfig{
		BaseURL:           getEnv("FRONTEND_BASE_URL", "http://localhost:3000"),
		PasswordResetPath: passwordResetPath,
	}, nil
}

package config

import (
	"fmt"
	"strings"
	"time"
)

type RateLimitConfig struct {
	LoginLimit     int
	LoginWindow    time.Duration
	RefreshLimit   int
	RefreshWindow  time.Duration
	LogoutLimit    int
	LogoutWindow   time.Duration
	TrustedProxies []string
}

func LoadRateLimitConfig() (*RateLimitConfig, error) {
	loginLimit, err := getIntEnv("RATE_LIMIT_LOGIN_MAX", 5)
	if err != nil {
		return nil, err
	}
	loginWindow, err := getTimeDurationEnv("RATE_LIMIT_LOGIN_WINDOW", 15*time.Minute)
	if err != nil {
		return nil, err
	}
	refreshLimit, err := getIntEnv("RATE_LIMIT_REFRESH_MAX", 10)
	if err != nil {
		return nil, err
	}
	refreshWindow, err := getTimeDurationEnv("RATE_LIMIT_REFRESH_WINDOW", time.Minute)
	if err != nil {
		return nil, err
	}
	logoutLimit, err := getIntEnv("RATE_LIMIT_LOGOUT_MAX", 20)
	if err != nil {
		return nil, err
	}
	logoutWindow, err := getTimeDurationEnv("RATE_LIMIT_LOGOUT_WINDOW", time.Minute)
	if err != nil {
		return nil, err
	}

	values := getEnv("RATE_LIMIT_TRUSTED_PROXIES", "")
	proxies := make([]string, 0)
	for _, value := range strings.Split(values, ",") {
		value = strings.TrimSpace(value)
		if value != "" {
			proxies = append(proxies, value)
		}
	}

	if loginLimit <= 0 || refreshLimit <= 0 || logoutLimit <= 0 {
		return nil, fmt.Errorf("rate limit maximums must be positive")
	}
	if loginWindow <= 0 || refreshWindow <= 0 || logoutWindow <= 0 {
		return nil, fmt.Errorf("rate limit windows must be positive")
	}

	return &RateLimitConfig{
		LoginLimit:     loginLimit,
		LoginWindow:    loginWindow,
		RefreshLimit:   refreshLimit,
		RefreshWindow:  refreshWindow,
		LogoutLimit:    logoutLimit,
		LogoutWindow:   logoutWindow,
		TrustedProxies: proxies,
	}, nil
}

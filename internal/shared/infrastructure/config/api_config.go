package config

type APIConfig struct {
	BaseURL string // e.g. "/api/v1" or "http://localhost:8080"
}

func LoadAPIConfig() (*APIConfig, error) {
	baseURL := getEnv("API_BASE_URL", "/api/v1")

	return &APIConfig{
		BaseURL: baseURL,
	}, nil
}
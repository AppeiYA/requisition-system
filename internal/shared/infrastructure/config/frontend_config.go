package config 

type FrontendConfig struct {
	BaseURL string
}

func LoadFrontendConfig() (*FrontendConfig, error) {
	return &FrontendConfig{
		BaseURL: getEnv("FRONTEND_BASE_URL", "http://localhost:3000"),
	}, nil
}
package config

import "fmt"

type AppBinaries int

const (
	OneBinary   AppBinaries = 1
	TwoBinaries AppBinaries = 2
)

func (b AppBinaries) IsValid() bool {
	return b == OneBinary || b == TwoBinaries
}

type Environment string

var (
	DevEnv     Environment = "development"
	TestEnv    Environment = "test"
	StagingEnv Environment = "staging"
	ProdEnv    Environment = "production"
)

func (e Environment) IsValid() bool {
	return e == DevEnv || e == StagingEnv || e == ProdEnv || e == TestEnv
}

type AppConfig struct {
	Env             Environment // "development", "staging", "production"
	Binaries        AppBinaries
	AnalyticsSource string
}

func LoadAppConfig() (*AppConfig, error) {
	env := getEnv("APP_ENV", "development")
	appEnv := Environment(env)
	if !appEnv.IsValid() {
		return nil, fmt.Errorf("invalid value for APP_ENV: %s", env)
	}

	binaries, err := getIntEnv("APP_BINARIES", 1)
	if err != nil {
		return nil, err
	}

	validBinary := AppBinaries(binaries)
	if !validBinary.IsValid() {
		return nil, fmt.Errorf("invalid value for APP_BINARIES: %d", binaries)
	}

	analyticsSource := getEnv("ANALYTICS_SOURCE", "direct")
	if analyticsSource != "direct" && analyticsSource != "materialized" {
		return nil, fmt.Errorf("invalid value for ANALYTICS_SOURCE: %s", analyticsSource)
	}

	return &AppConfig{
		Env:             appEnv,
		Binaries:        validBinary,
		AnalyticsSource: analyticsSource,
	}, nil
}

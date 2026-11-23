package config

type TestConfig struct {
	Database DBConfig
	Redis    RedisConfig
}

func LoadTestConfig() (*TestConfig, error) {
	dbConfig, err := LoadTestDBConfig()
	if err != nil {
		return nil, err
	}

	redisConfig, err := LoadTestRedisConfig()
	if err != nil {
		return nil, err
	}

	return &TestConfig{
		Database: *dbConfig,
		Redis:    *redisConfig,
	}, nil
}

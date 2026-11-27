package config

type Config struct {
	App           AppConfig
	API           APIConfig
	DB            DBConfig
	Redis         RedisConfig
	Server        ServerConfig
	JWT           JWTConfig
	Email         EmailConfig
	Frontend      FrontendConfig
	Worker        WorkerConfig
	Telemetry     TelemetryConfig
	RateLimit     RateLimitConfig
	PasswordReset PasswordResetConfig
	Attachment    AttachmentConfig
	Log           LogConfig
}

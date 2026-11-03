package redis

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

func Connect(cfg config.RedisConfig) (*Redis, error) {
	opts := &redis.Options{
		Addr:     fmt.Sprintf(`%s:%s`, cfg.Host, cfg.Port),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	if cfg.TLS {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		opts.TLSConfig = tlsConfig
	}

	client := redis.NewClient(opts)
	if err := redisotel.InstrumentTracing(client); err != nil {
		return nil, err
	}
	if err := redisotel.InstrumentMetrics(client); err != nil {
		return nil, err
	}

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return New(client), nil
}

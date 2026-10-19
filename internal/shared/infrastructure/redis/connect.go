package redis

import (
	"context"
	"fmt"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

func Connect(cfg config.RedisConfig) (*Redis, error) {
	opts := &redis.Options{
		Addr:     fmt.Sprintf(`%s:%s`, cfg.Host, cfg.Port),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	client := redis.NewClient(opts)

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return New(client), nil
}

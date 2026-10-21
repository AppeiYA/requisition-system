package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/cache"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) Get(
	ctx context.Context,
	key string,
	dest any,
) (bool, error) {
	data, err := c.client.Get(ctx, key).Bytes()

	if errors.Is(err, redis.Nil) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return false, err
	}

	return true, nil
}

func (c *Cache) Set(
	ctx context.Context,
	key string,
	value any,
	ttl time.Duration,
) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *Cache) Delete(
	ctx context.Context,
	key string,
) error {
	return c.client.Del(ctx, key).Err()
}

var _ cache.Cache = (*Cache)(nil)
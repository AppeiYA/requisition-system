package testutil

import (
	"context"
	"fmt"
	"time"

	redisclient "github.com/redis/go-redis/v9"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
)

const redisImage = "redis:7-alpine"

type TestRedis struct {
	Client    *redisclient.Client
	Container *rediscontainer.RedisContainer
}

func SetupTestRedis(ctx context.Context) (*TestRedis, error) {
	// 1. Start Redis container.
	container, err := rediscontainer.Run(
		ctx,
		redisImage,
	)
	if err != nil {
		return nil, fmt.Errorf("start redis container: %w", err)
	}

	// 2. Retrieve dynamically assigned connection details.
	connectionString, err := container.ConnectionString(ctx)
	if err != nil {
		_ = terminateRedisContainer(container)
		return nil, fmt.Errorf("get redis connection string: %w", err)
	}

	// 3. Parse the connection string.
	options, err := redisclient.ParseURL(connectionString)
	if err != nil {
		_ = terminateRedisContainer(container)
		return nil, fmt.Errorf("parse redis connection string: %w", err)
	}

	// 4. Create Redis client.
	client := redisclient.NewClient(options)

	// 5. Verify connectivity.
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		_ = terminateRedisContainer(container)

		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &TestRedis{
		Client:    client,
		Container: container,
	}, nil
}

// Close releases the Redis client and terminates its container.
func (r *TestRedis) Close() error {
	if r == nil {
		return nil
	}

	var clientErr error

	if r.Client != nil {
		clientErr = r.Client.Close()
	}

	var containerErr error

	if r.Container != nil {
		containerErr = terminateRedisContainer(r.Container)
	}

	if clientErr != nil {
		return fmt.Errorf("close redis client: %w", clientErr)
	}

	if containerErr != nil {
		return containerErr
	}

	return nil
}

func terminateRedisContainer(
	container *rediscontainer.RedisContainer,
) error {
	if container == nil {
		return nil
	}

	cleanupCtx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	if err := container.Terminate(
		cleanupCtx,
	); err != nil {
		return fmt.Errorf("terminate redis container: %w", err)
	}

	return nil
}

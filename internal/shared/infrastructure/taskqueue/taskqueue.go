package taskqueue

import (
	"context"
	"log"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/hibiken/asynq"
)

// NewClient returns an asynq client sharing the application's Redis
// connection pool. The pool is owned by the *redis.Redis it came from, so
// closing the client is neither needed nor allowed (asynq refuses to close
// a shared connection).
func NewClient(r *redis.Redis) *asynq.Client {
	return asynq.NewClientFromRedisClient(r.Client())
}

// NewServer returns the worker server. It reads tasks from the configured
// queue and logs any task error once asynq gives up on it.
func NewServer(r *redis.Redis, cfg config.WorkerConfig) *asynq.Server {
	return asynq.NewServerFromRedisClient(r.Client(), asynq.Config{
		Concurrency:     cfg.Concurrency,
		Queues:          map[string]int{cfg.Queue: 10},
		ShutdownTimeout: cfg.ShutdownTimeout,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			retried, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)
			log.Printf("task %s gave up after %d/%d retries: %v", task.Type(), retried, maxRetry, err)
		}),
	})
}

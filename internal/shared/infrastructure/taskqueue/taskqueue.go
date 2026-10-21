package taskqueue

import (
	"context"
	"log"
	"time"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/hibiken/asynq"
)

func NewClient(r *redis.Redis) *asynq.Client {
	return asynq.NewClientFromRedisClient(r.Client())
}

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

func NewScheduler(r *redis.Redis) *asynq.Scheduler {
	return asynq.NewSchedulerFromRedisClient(r.Client(), &asynq.SchedulerOpts{
		Location: time.UTC,
	})
}

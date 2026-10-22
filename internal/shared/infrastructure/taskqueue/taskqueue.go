package taskqueue

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/telemetry"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// traceParentField is the one field of a task payload the middleware
// understands: the traceparent the producer wrote. Reading it here instead of
// in each handler means every task type picks up the trace of the request
// that queued it by carrying the field, with no change to this middleware.
const traceParentField = "trace_parent"

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

// TraceMiddleware puts every task in the trace of the request that queued it.
// asynq hands the handler a context with the retry bookkeeping but without
// the producer's span, so it is read back from the payload here and the span
// of the task becomes a child of it.
//
// The span covers the handler, so the database and email calls it makes hang
// under the same trace. A task that carries no parent, the scheduled ones for
// instance, starts a trace of its own.
func TraceMiddleware() asynq.MiddlewareFunc {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
			ctx, span := telemetry.StartConsumer(ctx, "asynq "+task.Type(), traceParentOf(task))
			defer span.End()

			if id, ok := asynq.GetTaskID(ctx); ok {
				span.SetAttributes(attribute.String("asynq.task_id", id))
			}
			if queue, ok := asynq.GetQueueName(ctx); ok {
				span.SetAttributes(attribute.String("asynq.queue", queue))
			}
			if retried, ok := asynq.GetRetryCount(ctx); ok {
				span.SetAttributes(attribute.Int64("asynq.retry_count", int64(retried)))
			}
			if maxRetry, ok := asynq.GetMaxRetry(ctx); ok {
				span.SetAttributes(attribute.Int64("asynq.max_retry", int64(maxRetry)))
			}

			err := next.ProcessTask(ctx, task)
			if err == nil {
				return nil
			}

			telemetry.RecordError(span, err)

			// The task is about to be retried, so the span says how far the
			// retries have got. asynq is the one that decides whether the
			// error is retried at all, this only reports it.
			if retried, ok := asynq.GetRetryCount(ctx); ok {
				span.AddEvent("asynq.retry", trace.WithAttributes(
					attribute.Int64("asynq.retry_count", int64(retried)),
				))
			}

			return err
		})
	}
}

// traceParentOf reads the traceparent out of the payload. A payload that is
// not the JSON object this middleware expects has no parent to hand over,
// which is also the case for the scheduled tasks, whose payload is empty.
func traceParentOf(task *asynq.Task) string {
	var envelope struct {
		TraceParent string `json:"trace_parent"`
	}
	if err := json.Unmarshal(task.Payload(), &envelope); err != nil {
		return ""
	}
	return envelope.TraceParent
}

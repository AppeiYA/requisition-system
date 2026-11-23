package analyticsworker

import (
	"context"
	"log/slog"
	"time"

	analyticspostgres "github.com/AppeiYA/requisition-system/internal/analytics/adapters/out/postgres"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

const RefreshMaterializedViews = "analytics:refresh-materialized"
const refreshLockKey = "analytics:refresh"

type RefreshHandler struct {
	refresher *analyticspostgres.Refresher
	redis     *redis.Client
}

func NewRefreshHandler(refresher *analyticspostgres.Refresher, redisClient *redis.Client) *RefreshHandler {
	return &RefreshHandler{refresher: refresher, redis: redisClient}
}

func (h *RefreshHandler) Handle(ctx context.Context, _ *asynq.Task) error {
	acquired, err := h.redis.SetNX(ctx, refreshLockKey, "1", 10*time.Minute).Result()
	if err != nil {
		return err
	}
	if !acquired {
		return nil
	}
	defer func() {
		_ = h.redis.Del(context.Background(), refreshLockKey).Err()
	}()

	if err := h.refresher.Refresh(ctx); err != nil {
		slog.ErrorContext(ctx, "analytics materialized refresh failed", slog.String("error.message", err.Error()))
		return err
	}
	return nil
}

package analyticspostgres

import (
	"context"
	"fmt"
	"time"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
)

type Refresher struct {
	repository db.Repository
}

func NewRefresher(repository db.Repository) *Refresher {
	return &Refresher{repository: repository}
}

func (r *Refresher) Refresh(ctx context.Context) error {
	for _, viewName := range []string{"mv_requisition_facts", "mv_requisition_daily"} {
		startedAt := time.Now().UTC()
		if _, err := r.repository.Executor(ctx).ExecContext(ctx, `
			INSERT INTO analytics_refresh_state (view_name, started_at, last_error)
			VALUES ($1, $2, NULL)
			ON CONFLICT (view_name) DO UPDATE SET
			  started_at = EXCLUDED.started_at,
			  last_error = NULL`, viewName, startedAt); err != nil {
			return err
		}

		var isPopulated bool
		if err := r.repository.Executor(ctx).GetContext(ctx, &isPopulated, `
			SELECT ispopulated
			FROM pg_matviews
			WHERE schemaname = current_schema() AND matviewname = $1`, viewName); err != nil {
			r.recordFailure(ctx, viewName, startedAt, err)
			return err
		}

		query := "REFRESH MATERIALIZED VIEW " + viewName
		if isPopulated {
			query = "REFRESH MATERIALIZED VIEW CONCURRENTLY " + viewName
		}
		if _, err := r.repository.Executor(ctx).ExecContext(ctx, query); err != nil {
			r.recordFailure(ctx, viewName, startedAt, err)
			return err
		}

		var rowCount int64
		if err := r.repository.Executor(ctx).GetContext(ctx, &rowCount,
			"SELECT COUNT(*) FROM "+viewName); err != nil {
			r.recordFailure(ctx, viewName, startedAt, err)
			return err
		}
		refreshedAt := time.Now().UTC()
		duration := int(refreshedAt.Sub(startedAt).Milliseconds())
		if _, err := r.repository.Executor(ctx).ExecContext(ctx, `
			UPDATE analytics_refresh_state
			SET refreshed_at=$2, duration_ms=$3, row_count=$4, last_error=NULL
			WHERE view_name=$1`,
			viewName, refreshedAt, duration, rowCount); err != nil {
			return err
		}
	}
	return nil
}

func (r *Refresher) recordFailure(ctx context.Context, viewName string, startedAt time.Time, refreshErr error) {
	_, _ = r.repository.Executor(ctx).ExecContext(ctx, `
		UPDATE analytics_refresh_state
		SET duration_ms=$2, last_error=$3
		WHERE view_name=$1`,
		viewName,
		int(time.Now().UTC().Sub(startedAt).Milliseconds()),
		fmt.Sprintf("%v", refreshErr),
	)
}

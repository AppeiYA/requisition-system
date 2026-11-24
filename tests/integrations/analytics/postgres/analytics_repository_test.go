package postgres_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	analyticspostgres "github.com/AppeiYA/requisition-system/internal/analytics/adapters/out/postgres"
	analyticsdomain "github.com/AppeiYA/requisition-system/internal/analytics/domain"
	analyticsusecase "github.com/AppeiYA/requisition-system/internal/analytics/usecase"
	appdb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

func analyticsFilter(fixture *testutil.AnalyticsFixture) analyticsdomain.Filter {
	return analyticsdomain.Filter{
		From:     fixture.From,
		To:       fixture.To,
		Interval: "day",
		Limit:    10,
	}
}

func newAnalyticsRepositories(t *testing.T) (
	*testutil.Suite,
	*testutil.AnalyticsFixture,
	*analyticspostgres.Repository,
	*analyticspostgres.MaterializedRepository,
	*analyticspostgres.Refresher,
) {
	t.Helper()

	suite := testutil.NewSuite(t, "../../../../migrations")
	fixture := testutil.SeedAnalyticsFixture(t, suite)

	baseRepository := appdb.NewRepository(suite.AppDB)
	direct := analyticspostgres.NewRepository(baseRepository)
	materialized := analyticspostgres.NewMaterializedRepository(*direct)

	return suite, fixture, direct, materialized, analyticspostgres.NewRefresher(baseRepository)
}

func TestAnalyticsRepositoryDirectAggregates(t *testing.T) {
	suite, fixture, direct, _, _ := newAnalyticsRepositories(t)
	ctx := context.Background()
	filter := analyticsFilter(fixture)

	overview, err := direct.GetOverview(ctx, filter)
	require.NoError(t, err)
	require.Equal(t, 4, overview.Users.Total)
	require.Equal(t, 2, overview.Users.Active)
	require.Equal(t, 1, overview.Users.Suspended)
	require.Equal(t, 1, overview.Users.Inactive)
	require.Equal(t, 1, overview.Users.Admins)
	require.Equal(t, 3, overview.Users.NewInRange)
	require.Equal(t, 5, overview.Invitations.Total)
	require.Equal(t, 1, overview.Invitations.Accepted)
	require.InDelta(t, 0.2, overview.Invitations.AcceptanceRate, 1e-9)
	require.Equal(t, 5, overview.Requisitions.Total)
	require.Equal(t, map[string]int{
		"pending": 2, "approved": 1, "completed": 1, "rejected": 1,
	}, overview.Requisitions.ByStatus)
	require.Equal(t, 21000.0, overview.Requisitions.TotalEstimatedAmount)
	require.Equal(t, 4200.0, overview.Requisitions.AverageEstimatedAmount)
	require.Equal(t, 1, overview.Requisitions.PendingOverdue)
	require.Equal(t, 2, overview.Catalogue.ActiveItems)
	require.Equal(t, 1, overview.Catalogue.ActiveCategories)
	require.Equal(t, 2, overview.Catalogue.UnreviewedRequestedItems)

	byName := func(items []analyticsdomain.Breakdown) map[string]analyticsdomain.Breakdown {
		result := make(map[string]analyticsdomain.Breakdown, len(items))
		for _, item := range items {
			result[item.Name] = item
		}
		return result
	}

	requisitions, err := direct.GetRequisitionAnalytics(ctx, filter)
	require.NoError(t, err)
	require.Len(t, requisitions.Trend, 3)
	require.Equal(t, 3600.0, requisitions.CycleTime.PendingToApprovedSeconds)
	require.Equal(t, 7200.0, requisitions.CycleTime.ApprovedToCompletedSeconds)
	require.Equal(t, 14400.0, requisitions.CycleTime.CreationToClosedSeconds)
	require.Equal(t, 4, byName(requisitions.ByDepartment)["Engineering"].Count)
	require.Equal(t, 16000.0, byName(requisitions.ByDepartment)["Engineering"].EstimatedAmount)
	require.Equal(t, 3, byName(requisitions.ByCategory)["Office equipment"].Count)
	require.Equal(t, 15500.0, byName(requisitions.ByCategory)["Office equipment"].EstimatedAmount)
	require.Equal(t, 15000.0, byName(requisitions.TopItems)["Laptop"].EstimatedAmount)
	require.Equal(t, 17500.0, byName(requisitions.TopRequesters)["Ada Lovelace"].EstimatedAmount)

	users, err := direct.GetUserAnalytics(ctx, filter)
	require.NoError(t, err)
	require.Len(t, users.Growth, 3)
	roles := make(map[string]int, len(users.ByRole))
	for _, item := range users.ByRole {
		roles[item.Name] = item.Count
	}
	require.Equal(t, map[string]int{"user": 3, "admin": 1}, roles)
	require.Len(t, users.ByDepartment, 2)
	require.Equal(t, 5, users.InvitationFunnel.Created)
	require.InDelta(t, 0.2, users.InvitationFunnel.AcceptanceRate, 1e-9)
	require.InDelta(t, 172800.0, users.InvitationFunnel.AverageTimeToAcceptSeconds, 1e-6)

	catalogue, err := direct.GetCatalogueAnalytics(ctx, filter)
	require.NoError(t, err)
	require.Equal(t, 2, catalogue.ActiveItems)
	require.Equal(t, 1, catalogue.InactiveItems)
	require.Equal(t, 1, catalogue.ActiveCategories)
	require.Equal(t, 1, catalogue.InactiveCategories)
	require.Equal(t, 2, catalogue.RequestedItemReview.Unreviewed)
	require.Equal(t, 1, catalogue.RequestedItemReview.Promoted)
	require.Equal(t, 1, catalogue.RequestedItemReview.Dismissed)
	require.Equal(t, 2, catalogue.ReviewBacklogAging.OlderThan7Days)
	require.Equal(t, 1, catalogue.ReviewBacklogAging.OlderThan30Days)
	require.Equal(t, 3, byName(catalogue.TopCategories)["Office equipment"].Count)

	usecase := analyticsusecase.NewAnalyticsUsecase(direct, nil, "direct")
	withMetadata, err := usecase.ExecuteOverview(ctx, filter)
	require.NoError(t, err)
	require.Equal(t, "direct", withMetadata.Source)
	require.Equal(t, withMetadata.GeneratedAt, withMetadata.DataAsOf)

	_, err = usecase.ExecuteOverview(ctx, analyticsdomain.Filter{
		From:     fixture.To,
		To:       fixture.From,
		Interval: "day",
		Limit:    10,
	})
	require.ErrorIs(t, err, analyticsusecase.ErrInvalidRange)

	// No Redis is started for this suite, so the fixture must still be
	// queryable through the repository after the use case calls.
	var userCount int
	err = suite.DB.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&userCount)
	require.NoError(t, err)
	require.Equal(t, 4, userCount)
}

type refreshState struct {
	ViewName    string
	RefreshedAt sql.NullTime
	StartedAt   sql.NullTime
	DurationMS  sql.NullInt64
	RowCount    sql.NullInt64
	LastError   sql.NullString
}

func readRefreshState(t *testing.T, suite *testutil.Suite) map[string]refreshState {
	t.Helper()

	rows, err := suite.SQLX.QueryContext(context.Background(), `
		SELECT view_name, refreshed_at, started_at, duration_ms, row_count, last_error
		FROM analytics_refresh_state
	`)
	require.NoError(t, err)
	defer rows.Close()

	result := make(map[string]refreshState)
	for rows.Next() {
		var state refreshState
		require.NoError(t, rows.Scan(
			&state.ViewName,
			&state.RefreshedAt,
			&state.StartedAt,
			&state.DurationMS,
			&state.RowCount,
			&state.LastError,
		))
		result[state.ViewName] = state
	}
	require.NoError(t, rows.Err())
	return result
}

func TestAnalyticsMaterializedRefreshAndParity(t *testing.T) {
	suite, fixture, direct, materialized, refresher := newAnalyticsRepositories(t)
	ctx := context.Background()
	filter := analyticsFilter(fixture)

	t.Run("materialized views are unpopulated before a refresh", func(t *testing.T) {
		_, err := materialized.GetOverview(ctx, filter)
		require.ErrorContains(t, err, "has not been populated")
		require.Empty(t, readRefreshState(t, suite))
	})

	t.Run("refresh records state in facts then daily order", func(t *testing.T) {
		require.NoError(t, refresher.Refresh(ctx))

		state := readRefreshState(t, suite)
		require.Len(t, state, 2)
		for _, viewName := range []string{"mv_requisition_facts", "mv_requisition_daily"} {
			row := state[viewName]
			require.True(t, row.RefreshedAt.Valid)
			require.True(t, row.StartedAt.Valid)
			require.True(t, row.DurationMS.Valid)
			require.GreaterOrEqual(t, row.DurationMS.Int64, int64(0))
			require.True(t, row.RowCount.Valid)
			require.Equal(t, int64(6), row.RowCount.Int64)
			require.False(t, row.LastError.Valid)
		}
		require.False(t,
			state["mv_requisition_facts"].StartedAt.Time.After(
				state["mv_requisition_daily"].StartedAt.Time,
			),
		)
	})

	t.Run("direct and materialized aggregates agree after refresh", func(t *testing.T) {
		directOverview, err := direct.GetOverview(ctx, filter)
		require.NoError(t, err)
		materializedOverview, err := materialized.GetOverview(ctx, filter)
		require.NoError(t, err)

		require.Equal(t, directOverview.Requisitions.Total, materializedOverview.Requisitions.Total)
		require.Equal(t, directOverview.Requisitions.ByStatus, materializedOverview.Requisitions.ByStatus)
		require.Equal(t, directOverview.Requisitions.TotalEstimatedAmount, materializedOverview.Requisitions.TotalEstimatedAmount)
		require.Equal(t, directOverview.Requisitions.AverageEstimatedAmount, materializedOverview.Requisitions.AverageEstimatedAmount)
		require.Equal(t, directOverview.Requisitions.PendingOverdue, materializedOverview.Requisitions.PendingOverdue)

		directRequisitions, err := direct.GetRequisitionAnalytics(ctx, filter)
		require.NoError(t, err)
		materializedRequisitions, err := materialized.GetRequisitionAnalytics(ctx, filter)
		require.NoError(t, err)

		require.ElementsMatch(t, directRequisitions.ByStatus, materializedRequisitions.ByStatus)
		require.ElementsMatch(t, directRequisitions.ByPriority, materializedRequisitions.ByPriority)
		require.Equal(t, directRequisitions.ByDepartment, materializedRequisitions.ByDepartment)

		require.Equal(t, directRequisitions.Trend, materializedRequisitions.Trend)
		// Sections that remain live are identical by construction.
		require.Equal(t, directRequisitions.ByCategory, materializedRequisitions.ByCategory)
		require.Equal(t, directRequisitions.CycleTime, materializedRequisitions.CycleTime)

		require.Equal(t, "materialized", materializedOverview.Source)
		require.False(t, materializedOverview.DataAsOf.IsZero())
	})

	t.Run("data_as_of is the oldest dependent refresh", func(t *testing.T) {
		older := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Microsecond)
		_, err := suite.DB.Pool.Exec(ctx, `
			UPDATE analytics_refresh_state
			SET refreshed_at = $1
			WHERE view_name = 'mv_requisition_facts'
		`, older)
		require.NoError(t, err)

		overview, err := materialized.GetOverview(ctx, filter)
		require.NoError(t, err)
		require.WithinDuration(t, older, overview.DataAsOf, time.Millisecond)
	})

	t.Run("a second refresh uses concurrently and includes new rows", func(t *testing.T) {
		tx, err := suite.SQLX.BeginTxx(ctx, nil)
		require.NoError(t, err)
		defer func() { _ = tx.Rollback() }()

		createdAt := fixture.Anchor.AddDate(0, 0, -1).Add(12 * time.Hour)
		_, err = tx.ExecContext(ctx, `
			INSERT INTO requisitions (
				id, reference_code, requested_by, department_id, needed_by,
				justification, priority, status, created_at, updated_at
			) VALUES (
				'req-analytics-refresh', 'REQ-ANALYTICS-REFRESH', $1, $2, $3,
				'Refresh fixture', 'urgent', 'pending', $4, $4
			)
		`, fixture.UserID, testutil.AnalyticsEngineeringDepartmentID,
			createdAt.AddDate(0, 0, 3).Format("2006-01-02"), createdAt)
		require.NoError(t, err)

		_, err = tx.ExecContext(ctx, `
			INSERT INTO requisition_status_history (
				id, requisition_id, sequence, from_status, to_status,
				acted_by, actor_role, created_at
			) VALUES (
				'rsh-analytics-refresh', 'req-analytics-refresh', 1, NULL, 'pending',
				$1, 'staff', $2
			)
		`, fixture.UserID, createdAt)
		require.NoError(t, err)
		require.NoError(t, tx.Commit())

		// The views are populated now, so this refresh takes the
		// REFRESH MATERIALIZED VIEW CONCURRENTLY path.
		require.NoError(t, refresher.Refresh(ctx))

		state := readRefreshState(t, suite)
		require.Equal(t, int64(7), state["mv_requisition_facts"].RowCount.Int64)
		require.Equal(t, int64(7), state["mv_requisition_daily"].RowCount.Int64)
		for _, row := range state {
			require.False(t, row.LastError.Valid, row.LastError.String)
		}

		overview, err := materialized.GetOverview(ctx, filter)
		require.NoError(t, err)
		require.Equal(t, 6, overview.Requisitions.Total)
		require.Equal(t, 3, overview.Requisitions.ByStatus["pending"])
	})
}

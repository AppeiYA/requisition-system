package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	analyticshttp "github.com/AppeiYA/requisition-system/internal/analytics/adapters/in/http"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const analyticsBasePath = "/api/v1/admin/analytics"

var analyticsEndpoints = []string{"overview", "requisitions", "users", "catalogue"}

type analyticsEnvelope[T any] struct {
	Success bool                  `json:"success"`
	Data    T                     `json:"data"`
	Error   string                `json:"error"`
	Details []response.FieldError `json:"details"`
}

func analyticsRequest(router *gin.Engine, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func analyticsToken(
	t *testing.T,
	service identityout.AccessTokenService,
	userID,
	role string,
) string {
	t.Helper()

	token, err := service.GenerateAccessToken(
		context.Background(),
		&identityout.TokenModel{ID: userID, Role: role},
	)
	require.NoError(t, err)
	return token.Value
}

func decodeAnalytics[T any](t *testing.T, rec *httptest.ResponseRecorder) analyticsEnvelope[T] {
	t.Helper()

	var body analyticsEnvelope[T]
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), rec.Body.String())
	require.True(t, body.Success, rec.Body.String())
	return body
}

func fixtureQuery(fixture *testutil.AnalyticsFixture) string {
	values := url.Values{}
	values.Set("from", fixture.From.Format(time.RFC3339))
	values.Set("to", fixture.To.Format(time.RFC3339))
	return "?" + values.Encode()
}

func breakdownsByName(items []analyticshttp.BreakdownResponse) map[string]analyticshttp.BreakdownResponse {
	result := make(map[string]analyticshttp.BreakdownResponse, len(items))
	for _, item := range items {
		result[item.Name] = item
	}
	return result
}

func countValuesByName(items []analyticshttp.CountValueResponse) map[string]int {
	result := make(map[string]int, len(items))
	for _, item := range items {
		result[item.Name] = item.Count
	}
	return result
}

func trendByBucket(items []analyticshttp.TrendResponse) map[string]analyticshttp.TrendResponse {
	result := make(map[string]analyticshttp.TrendResponse, len(items))
	for _, item := range items {
		result[item.Bucket.Format(time.RFC3339)] = item
	}
	return result
}

func TestAnalyticsHTTPDirect(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	analytics := suite.SetupAnalyticsHTTP(t, "direct")
	fixture := testutil.SeedAnalyticsFixture(t, suite)

	adminToken := analyticsToken(t, analytics.AccessTokens, fixture.AdminID, "admin")
	userToken := analyticsToken(t, analytics.AccessTokens, fixture.UserID, "user")

	t.Run("requires a valid admin token on every endpoint", func(t *testing.T) {
		for _, endpoint := range analyticsEndpoints {
			path := analyticsBasePath + "/" + endpoint

			rec := analyticsRequest(analytics.Router, path, "")
			require.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())

			rec = analyticsRequest(analytics.Router, path, userToken)
			require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
		}
	})

	t.Run("overview matches the fixed fixture", func(t *testing.T) {
		rec := analyticsRequest(
			analytics.Router,
			analyticsBasePath+"/overview"+fixtureQuery(fixture),
			adminToken,
		)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		body := decodeAnalytics[analyticshttp.OverviewResponse](t, rec)
		data := body.Data

		require.Equal(t, "direct", data.Source)
		require.Equal(t, data.GeneratedAt, data.DataAsOf)
		require.Equal(t, fixture.From.UTC(), data.Range.From.UTC())
		require.Equal(t, fixture.To.UTC(), data.Range.To.UTC())

		require.Equal(t, 4, data.Users.Total)
		require.Equal(t, 2, data.Users.Active)
		require.Equal(t, 1, data.Users.Suspended)
		require.Equal(t, 1, data.Users.Inactive)
		require.Equal(t, 1, data.Users.Admins)
		require.Equal(t, 3, data.Users.NewInRange)

		require.Equal(t, 5, data.Invitations.Total)
		require.Equal(t, 1, data.Invitations.Pending)
		require.Equal(t, 1, data.Invitations.Delivered)
		require.Equal(t, 1, data.Invitations.Accepted)
		require.Equal(t, 1, data.Invitations.Expired)
		require.Equal(t, 0, data.Invitations.Revoked)
		require.Equal(t, 1, data.Invitations.Failed)
		require.InDelta(t, 0.2, data.Invitations.AcceptanceRate, 1e-9)

		require.Equal(t, 5, data.Requisitions.Total)
		require.Equal(t, map[string]int{
			"pending": 2, "approved": 1, "completed": 1, "rejected": 1,
		}, data.Requisitions.ByStatus)
		require.Equal(t, 21000.0, data.Requisitions.TotalEstimatedAmount)
		require.Equal(t, 4200.0, data.Requisitions.AverageEstimatedAmount)
		require.Equal(t, "NGN", data.Requisitions.Currency)
		require.Equal(t, 1, data.Requisitions.PendingOverdue)

		require.Equal(t, 2, data.Catalogue.ActiveItems)
		require.Equal(t, 1, data.Catalogue.ActiveCategories)
		require.Equal(t, 2, data.Catalogue.UnreviewedRequestedItems)
	})

	t.Run("rejects invalid filters with field details", func(t *testing.T) {
		cases := []struct {
			name  string
			query string
			field string
		}{
			{"from after to", "?from=2026-01-02&to=2026-01-01", "from"},
			{"from equal to", "?from=2026-01-01&to=2026-01-01", "from"},
			{"range over 366 days", "?from=2024-01-01&to=2026-01-01", "to"},
			{"invalid interval", "?interval=hour", "interval"},
			{"invalid limit", "?limit=51", "limit"},
			{"invalid department", "?department_id=not-an-id", "department_id"},
			{"unparseable date", "?from=not-a-date", "from"},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				rec := analyticsRequest(
					analytics.Router,
					analyticsBasePath+"/overview"+tc.query,
					adminToken,
				)
				require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())

				var body analyticsEnvelope[any]
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				require.False(t, body.Success)
				require.Equal(t, "validation_error", body.Error)
				require.NotEmpty(t, body.Details)
				require.Equal(t, tc.field, body.Details[0].Field)
			})
		}
	})

	t.Run("scopes overview requisitions and invitations by department", func(t *testing.T) {
		query := fixtureQuery(fixture) + "&department_id=" + testutil.AnalyticsFinanceDepartmentID
		rec := analyticsRequest(analytics.Router, analyticsBasePath+"/overview"+query, adminToken)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		data := decodeAnalytics[analyticshttp.OverviewResponse](t, rec).Data
		require.Equal(t, 2, data.Invitations.Total)
		require.Equal(t, 1, data.Invitations.Pending)
		require.Equal(t, 1, data.Invitations.Expired)
		require.Equal(t, 1, data.Requisitions.Total)
		require.Equal(t, map[string]int{"completed": 1}, data.Requisitions.ByStatus)
		require.Equal(t, 5000.0, data.Requisitions.TotalEstimatedAmount)
	})

	t.Run("requisition analytics matches the fixed fixture", func(t *testing.T) {
		rec := analyticsRequest(
			analytics.Router,
			analyticsBasePath+"/requisitions"+fixtureQuery(fixture),
			adminToken,
		)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		data := decodeAnalytics[analyticshttp.RequisitionAnalyticsResponse](t, rec).Data

		byStatus := breakdownsByName(data.ByStatus)
		require.Equal(t, 2, byStatus["pending"].Count)
		require.Equal(t, 15500.0, byStatus["pending"].EstimatedAmount)
		require.Equal(t, 1, byStatus["approved"].Count)
		require.Equal(t, 500.0, byStatus["approved"].EstimatedAmount)
		require.Equal(t, 1, byStatus["completed"].Count)
		require.Equal(t, 5000.0, byStatus["completed"].EstimatedAmount)
		require.Equal(t, 1, byStatus["rejected"].Count)

		byPriority := breakdownsByName(data.ByPriority)
		require.Equal(t, 2, byPriority["high"].Count)
		require.Equal(t, 17500.0, byPriority["high"].EstimatedAmount)
		require.Equal(t, 1, byPriority["medium"].Count)
		require.Equal(t, 2, byPriority["low"].Count)
		require.Equal(t, 3000.0, byPriority["low"].EstimatedAmount)

		byDepartment := breakdownsByName(data.ByDepartment)
		require.Equal(t, 4, byDepartment["Engineering"].Count)
		require.Equal(t, 16000.0, byDepartment["Engineering"].EstimatedAmount)
		require.Equal(t, 1, byDepartment["Finance"].Count)
		require.Equal(t, 5000.0, byDepartment["Finance"].EstimatedAmount)

		byCategory := breakdownsByName(data.ByCategory)
		require.Equal(t, 3, byCategory["Office equipment"].Count)
		require.Equal(t, 15500.0, byCategory["Office equipment"].EstimatedAmount)
		require.Equal(t, 2, byCategory["Furniture"].Count)
		require.Equal(t, 5500.0, byCategory["Furniture"].EstimatedAmount)

		trend := trendByBucket(data.Trend)
		require.Len(t, trend, 3)
		require.Equal(t, 1, trend[fixture.Anchor.AddDate(0, 0, -3).Format(time.RFC3339)].Count)
		require.Equal(t, 5000.0, trend[fixture.Anchor.AddDate(0, 0, -3).Format(time.RFC3339)].EstimatedAmount)
		require.Equal(t, 2, trend[fixture.Anchor.AddDate(0, 0, -2).Format(time.RFC3339)].Count)
		require.Equal(t, 12500.0, trend[fixture.Anchor.AddDate(0, 0, -2).Format(time.RFC3339)].EstimatedAmount)
		require.Equal(t, 2, trend[fixture.Anchor.AddDate(0, 0, -1).Format(time.RFC3339)].Count)
		require.Equal(t, 3500.0, trend[fixture.Anchor.AddDate(0, 0, -1).Format(time.RFC3339)].EstimatedAmount)

		require.Equal(t, 3600.0, data.CycleTime.PendingToApprovedSeconds)
		require.Equal(t, 7200.0, data.CycleTime.ApprovedToCompletedSeconds)
		require.Equal(t, 14400.0, data.CycleTime.CreationToClosedSeconds)

		topItems := breakdownsByName(data.TopItems)
		require.Equal(t, 5, topItems["Notebook"].Count)
		require.Equal(t, 500.0, topItems["Notebook"].EstimatedAmount)
		require.Equal(t, 3, topItems["Laptop"].Count)
		require.Equal(t, 15000.0, topItems["Laptop"].EstimatedAmount)
		require.Equal(t, 2, topItems["Ergonomic chair"].Count)
		require.Equal(t, 1, topItems["Standing desk"].Count)

		topRequesters := breakdownsByName(data.TopRequesters)
		require.Equal(t, 3, topRequesters["Grace Hopper"].Count)
		require.Equal(t, 2, topRequesters["Ada Lovelace"].Count)
		require.Equal(t, 17500.0, topRequesters["Ada Lovelace"].EstimatedAmount)
	})

	t.Run("user analytics matches the fixed fixture", func(t *testing.T) {
		rec := analyticsRequest(
			analytics.Router,
			analyticsBasePath+"/users"+fixtureQuery(fixture),
			adminToken,
		)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		data := decodeAnalytics[analyticshttp.UserAnalyticsResponse](t, rec).Data

		growth := trendByBucket(data.Growth)
		require.Len(t, growth, 3)
		require.Equal(t, 1, growth[fixture.Anchor.AddDate(0, 0, -10).Format(time.RFC3339)].Count)
		require.Equal(t, 1, growth[fixture.Anchor.AddDate(0, 0, -5).Format(time.RFC3339)].Count)
		require.Equal(t, 1, growth[fixture.Anchor.AddDate(0, 0, -4).Format(time.RFC3339)].Count)

		require.Equal(t, map[string]int{"user": 3, "admin": 1}, countValuesByName(data.ByRole))
		require.Equal(t, map[string]int{
			"active": 2, "suspended": 1, "inactive": 1,
		}, countValuesByName(data.ByStatus))

		require.Len(t, data.ByDepartment, 2)
		for _, department := range data.ByDepartment {
			switch department.Name {
			case "Engineering":
				require.Equal(t, 2, department.Members)
				require.Equal(t, 2, department.ActiveMembers)
			case "Finance":
				require.Equal(t, 1, department.Members)
				require.Equal(t, 0, department.ActiveMembers)
			default:
				t.Fatalf("unexpected department %q", department.Name)
			}
		}

		funnel := data.InvitationFunnel
		require.Equal(t, 5, funnel.Created)
		require.Equal(t, 1, funnel.Delivered)
		require.Equal(t, 1, funnel.Accepted)
		require.Equal(t, 1, funnel.Expired)
		require.Equal(t, 0, funnel.Revoked)
		require.Equal(t, 1, funnel.Failed)
		require.InDelta(t, 0.2, funnel.AcceptanceRate, 1e-9)
		require.InDelta(t, 172800.0, funnel.AverageTimeToAcceptSeconds, 1e-6)

		topRequesters := breakdownsByName(data.TopRequesters)
		require.Equal(t, 3, topRequesters["Grace Hopper"].Count)
		require.Equal(t, 3500.0, topRequesters["Grace Hopper"].EstimatedAmount)
		require.Equal(t, 2, topRequesters["Ada Lovelace"].Count)
	})

	t.Run("catalogue analytics matches the fixed fixture", func(t *testing.T) {
		rec := analyticsRequest(
			analytics.Router,
			analyticsBasePath+"/catalogue"+fixtureQuery(fixture),
			adminToken,
		)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		data := decodeAnalytics[analyticshttp.CatalogueAnalyticsResponse](t, rec).Data
		require.Equal(t, 2, data.ActiveItems)
		require.Equal(t, 1, data.InactiveItems)
		require.Equal(t, 1, data.ActiveCategories)
		require.Equal(t, 1, data.InactiveCategories)
		require.Equal(t, 2, data.RequestedItemReview.Unreviewed)
		require.Equal(t, 1, data.RequestedItemReview.Promoted)
		require.Equal(t, 1, data.RequestedItemReview.Dismissed)
		require.Equal(t, 2, data.ReviewBacklogAging.OlderThan7Days)
		require.Equal(t, 1, data.ReviewBacklogAging.OlderThan30Days)

		topItems := breakdownsByName(data.TopItems)
		require.Equal(t, 5, topItems["Notebook"].Count)
		require.Equal(t, 3, topItems["Laptop"].Count)
		require.Equal(t, 15000.0, topItems["Laptop"].EstimatedAmount)

		topCategories := breakdownsByName(data.TopCategories)
		require.Equal(t, 3, topCategories["Office equipment"].Count)
		require.Equal(t, 2, topCategories["Furniture"].Count)
	})

	t.Run("caches responses by endpoint and filter", func(t *testing.T) {
		ctx := context.Background()
		path := analyticsBasePath + "/overview" + fixtureQuery(fixture)

		require.NoError(t, suite.ResetRedis(ctx))

		first := decodeAnalytics[analyticshttp.OverviewResponse](
			t,
			analyticsRequest(analytics.Router, path, adminToken),
		).Data
		keys, err := suite.Redis.Client.Keys(ctx, "analytics:v1:direct:overview:*").Result()
		require.NoError(t, err)
		require.Len(t, keys, 1)

		second := decodeAnalytics[analyticshttp.OverviewResponse](
			t,
			analyticsRequest(analytics.Router, path, adminToken),
		).Data
		require.Equal(t, first.GeneratedAt, second.GeneratedAt)

		require.NoError(t, suite.ResetRedis(ctx))
		keys, err = suite.Redis.Client.Keys(ctx, "analytics:v1:direct:*").Result()
		require.NoError(t, err)
		require.Empty(t, keys)

		third := decodeAnalytics[analyticshttp.OverviewResponse](
			t,
			analyticsRequest(analytics.Router, path, adminToken),
		).Data
		require.NotEqual(t, first.GeneratedAt, third.GeneratedAt)

		keys, err = suite.Redis.Client.Keys(ctx, "analytics:v1:direct:overview:*").Result()
		require.NoError(t, err)
		require.Len(t, keys, 1)
	})
}

func TestAnalyticsHTTPMaterializedSource(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	analytics := suite.SetupAnalyticsHTTP(t, "materialized")
	fixture := testutil.SeedAnalyticsFixture(t, suite)
	ctx := context.Background()

	require.NoError(t, analytics.Refresher.Refresh(ctx))
	adminToken := analyticsToken(t, analytics.AccessTokens, fixture.AdminID, "admin")

	rec := analyticsRequest(
		analytics.Router,
		analyticsBasePath+"/overview"+fixtureQuery(fixture),
		adminToken,
	)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	data := decodeAnalytics[analyticshttp.OverviewResponse](t, rec).Data
	require.Equal(t, "materialized", data.Source)
	require.Equal(t, 5, data.Requisitions.Total)
	require.Equal(t, 21000.0, data.Requisitions.TotalEstimatedAmount)

	var refreshedAt time.Time
	err := suite.DB.Pool.QueryRow(ctx, `
		SELECT MIN(refreshed_at)
		FROM analytics_refresh_state
		WHERE view_name IN ('mv_requisition_facts', 'mv_requisition_daily')
	`).Scan(&refreshedAt)
	require.NoError(t, err)
	require.False(t, refreshedAt.IsZero())
	require.WithinDuration(t, refreshedAt, data.DataAsOf, time.Millisecond)
}

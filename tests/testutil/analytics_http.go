package testutil

import (
	"testing"

	analyticshttp "github.com/AppeiYA/requisition-system/internal/analytics/adapters/in/http"
	analyticspostgres "github.com/AppeiYA/requisition-system/internal/analytics/adapters/out/postgres"
	analyticsin "github.com/AppeiYA/requisition-system/internal/analytics/ports/in"
	analyticsout "github.com/AppeiYA/requisition-system/internal/analytics/ports/out"
	analyticsusecase "github.com/AppeiYA/requisition-system/internal/analytics/usecase"
	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	appdb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	redisinfra "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
)

// AnalyticsHTTP wires the analytics handler into the admin HTTP harness so
// integration tests exercise the same auth middleware as the running app.
type AnalyticsHTTP struct {
	*AdminHTTP
	DirectRepository       *analyticspostgres.Repository
	MaterializedRepository *analyticspostgres.MaterializedRepository
	Refresher              *analyticspostgres.Refresher
}

// SetupAnalyticsHTTP starts the admin harness and registers the analytics
// routes with source set to "direct" or "materialized". Call it once per
// suite because SetupAdminHTTP resets the database and Redis.
func (s *Suite) SetupAnalyticsHTTP(t *testing.T, source string) *AnalyticsHTTP {
	t.Helper()

	admin := s.SetupAdminHTTP(t)
	if s.Redis == nil {
		t.Fatal("analytics HTTP setup requires Redis")
	}

	baseRepository := appdb.NewRepository(s.AppDB)
	directRepository := analyticspostgres.NewRepository(baseRepository)
	materializedRepository := analyticspostgres.NewMaterializedRepository(*directRepository)

	var repository analyticsout.Repository = directRepository
	if source == "materialized" {
		repository = materializedRepository
	}

	usecase := analyticsusecase.NewAnalyticsUsecase(
		repository,
		redisinfra.NewCache(s.Redis.Client),
		source,
	)
	portIn := &analyticsin.PortIn{
		Overview:     usecase,
		Requisitions: usecase,
		Users:        usecase,
		Catalogue:    usecase,
	}

	authMiddleware := identityhttp.NewAuthMiddleware(admin.AccessTokens, admin.UserRepository)
	handler := analyticshttp.NewHandler(portIn)
	handler.Register(
		admin.Router.Group("/api/v1"),
		authMiddleware.Handle(),
		authMiddleware.RequireRole("admin"),
	)

	return &AnalyticsHTTP{
		AdminHTTP:              admin,
		DirectRepository:       directRepository,
		MaterializedRepository: materializedRepository,
		Refresher:              analyticspostgres.NewRefresher(baseRepository),
	}
}

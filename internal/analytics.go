package app

import (
	analyticsmodule "github.com/AppeiYA/requisition-system/internal/analytics"
	analyticshhttp "github.com/AppeiYA/requisition-system/internal/analytics/adapters/in/http"
	analyticspostgres "github.com/AppeiYA/requisition-system/internal/analytics/adapters/out/postgres"
	analyticsin "github.com/AppeiYA/requisition-system/internal/analytics/ports/in"
	analyticsout "github.com/AppeiYA/requisition-system/internal/analytics/ports/out"
	analyticsusecase "github.com/AppeiYA/requisition-system/internal/analytics/usecase"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/cache"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
)

func buildAnalytics(baseRepo db.Repository, cache cache.Cache, source string) *analyticsmodule.Module {
	directRepository := *analyticspostgres.NewRepository(baseRepo)
	var repository analyticsout.Repository = &directRepository
	if source == "materialized" {
		repository = analyticspostgres.NewMaterializedRepository(directRepository)
	}
	usecase := analyticsusecase.NewAnalyticsUsecase(repository, cache, source)
	portIn := &analyticsin.PortIn{
		Overview:     usecase,
		Requisitions: usecase,
		Users:        usecase,
		Catalogue:    usecase,
	}
	handler := analyticshhttp.NewHandler(portIn)
	return analyticsmodule.NewModule(handler)
}

package in

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/analytics/domain"
)

type FilterUseCase interface {
	Execute(ctx context.Context, filter domain.Filter) (domain.Filter, error)
}

type GetOverviewUseCase interface {
	ExecuteOverview(ctx context.Context, filter domain.Filter) (*domain.Overview, error)
}

type GetRequisitionsUseCase interface {
	ExecuteRequisitions(ctx context.Context, filter domain.Filter) (*domain.RequisitionAnalytics, error)
}

type GetUsersUseCase interface {
	ExecuteUsers(ctx context.Context, filter domain.Filter) (*domain.UserAnalytics, error)
}

type GetCatalogueUseCase interface {
	ExecuteCatalogue(ctx context.Context, filter domain.Filter) (*domain.CatalogueAnalytics, error)
}

type PortIn struct {
	Overview     GetOverviewUseCase
	Requisitions GetRequisitionsUseCase
	Users        GetUsersUseCase
	Catalogue    GetCatalogueUseCase
}

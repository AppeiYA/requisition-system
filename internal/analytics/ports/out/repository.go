package out

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/analytics/domain"
)

type Repository interface {
	GetOverview(ctx context.Context, filter domain.Filter) (*domain.Overview, error)
	GetRequisitionAnalytics(ctx context.Context, filter domain.Filter) (*domain.RequisitionAnalytics, error)
	GetUserAnalytics(ctx context.Context, filter domain.Filter) (*domain.UserAnalytics, error)
	GetCatalogueAnalytics(ctx context.Context, filter domain.Filter) (*domain.CatalogueAnalytics, error)
}

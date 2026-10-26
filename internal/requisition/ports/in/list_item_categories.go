package requisitionin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type ListItemCategoriesQuery struct {
	Limit  int
	Offset int
}

type ListItemCategoriesUsecaseInt interface {
	Execute(ctx context.Context, query ListItemCategoriesQuery) ([]*domain.ItemCategory, error)
}

package categoryin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type ListItemsInCategoryUsecaseInt interface {
	Execute(ctx context.Context, categoryID string) ([]*domain.Item, error)
}

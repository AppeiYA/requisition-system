package categoryin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type DeactivateItemCategoryUsecaseInt interface {
	Execute(ctx context.Context, id string) (*domain.ItemCategory, error)
}

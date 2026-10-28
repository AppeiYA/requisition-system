package categoryin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type ActivateItemCategoryUsecaseInt interface {
	Execute(ctx context.Context, id string) (*domain.ItemCategory, error)
}

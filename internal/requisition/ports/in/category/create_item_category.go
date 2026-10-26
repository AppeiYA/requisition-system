package categoryin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

// CreateItemCategoryCommand carries the admin-provided details of a new item
// category.
type CreateItemCategoryCommand struct {
	Name        string
	Description string
}

type CreateItemCategoryUsecaseInt interface {
	Execute(ctx context.Context, cmd CreateItemCategoryCommand) (*domain.ItemCategory, error)
}

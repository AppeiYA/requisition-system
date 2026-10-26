package categoryin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

// UpdateItemCategoryCommand carries a partial update. A nil field is left
// unchanged, and the domain requires at least one field to be provided.
type UpdateItemCategoryCommand struct {
	ID          string
	Name        *string
	Description *string
}

type UpdateItemCategoryUsecaseInt interface {
	Execute(ctx context.Context, cmd UpdateItemCategoryCommand) (*domain.ItemCategory, error)
}

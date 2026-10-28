package itemin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

// UpdateItemCommand carries a partial update. A nil field is left unchanged,
// and the domain requires at least one field to be provided.
type UpdateItemCommand struct {
	ID          string
	Name        *string
	Description *string
	CategoryID  *string
	UnitCost    *float64
}

type UpdateItemUsecaseInt interface {
	Execute(ctx context.Context, cmd UpdateItemCommand) (*domain.Item, error)
}

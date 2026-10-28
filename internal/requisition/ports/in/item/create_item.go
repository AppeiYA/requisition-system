package itemin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

// CreateItemCommand carries the admin-provided details of a new item.
type CreateItemCommand struct {
	Name        string
	Description string
	CategoryID  string
	UnitCost    float64
}

type CreateItemUsecaseInt interface {
	Execute(ctx context.Context, cmd CreateItemCommand) (*domain.Item, error)
}

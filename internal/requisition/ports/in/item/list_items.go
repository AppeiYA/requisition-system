package itemin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type ListItemsQuery struct {
	Limit  int
	Offset int
}

type ListItemsUsecaseInt interface {
	Execute(ctx context.Context, query ListItemsQuery) ([]*domain.Item, error)
}

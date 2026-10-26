package requisitionout

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type ItemRepository interface {
	SaveItem(ctx context.Context, item *domain.Item) error
	GetItemByID(ctx context.Context, id string) (*domain.Item, error)
	ListItems(ctx context.Context, limit int, offset int) ([]*domain.Item, error)
	UpdateItem(ctx context.Context, item *domain.Item) error
}

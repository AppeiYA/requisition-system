package requisitionout

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type RequestedItemRepository interface {
	SaveRequestedItem(ctx context.Context, item *domain.RequestedItem) error
	GetRequestedItemByID(ctx context.Context, id string) (*domain.RequestedItem, error)
	GetRequestedItemByIDForUpdate(ctx context.Context, id string) (*domain.RequestedItem, error)
	ListUnreviewedRequestedItems(ctx context.Context, limit int, offset int) ([]*domain.RequestedItem, error)
	UpdateRequestedItemStatus(ctx context.Context, item *domain.RequestedItem) error
}

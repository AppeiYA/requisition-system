package requisitionout

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type RequestedItemRepository interface {
	SaveRequestedItem(ctx context.Context, item *domain.RequestedItem) error
	GetRequestedItemByID(ctx context.Context, id string) (*domain.RequestedItem, error)
}

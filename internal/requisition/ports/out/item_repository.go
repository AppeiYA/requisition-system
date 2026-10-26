package requisitionout

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type RequisitionRepository interface {
	SaveItem(ctx context.Context, item *domain.Item) error
	FindItemByID(ctx context.Context, id string) (*domain.Item, error)
	ListItems(ctx context.Context, limit int, offset int) ([]*domain.Item, error)
}

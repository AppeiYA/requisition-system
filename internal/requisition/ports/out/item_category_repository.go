package requisitionout

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type ItemCategoryRepository interface {
	SaveItemCategory(ctx context.Context, category *domain.ItemCategory) error
	GetItemCategoryByID(ctx context.Context, id string) (*domain.ItemCategory, error)
	ListItemCategories(ctx context.Context, limit int, offset int) ([]*domain.ItemCategory, error)
	UpdateItemCategory(ctx context.Context, category *domain.ItemCategory) error
}

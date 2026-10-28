package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

type ListItemsUsecase struct {
	itemRepo requisitionout.ItemRepository
}

func NewListItemsUsecase(
	itemRepo requisitionout.ItemRepository,
) *ListItemsUsecase {
	return &ListItemsUsecase{
		itemRepo: itemRepo,
	}
}

func (uc *ListItemsUsecase) Execute(
	ctx context.Context,
	query itemin.ListItemsQuery,
) ([]*domain.Item, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	limit, offset := shareddomain.Clamp(query.Limit, query.Offset)
	return uc.itemRepo.ListItems(ctx, limit, offset)
}

package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

type ActivateItemUsecase struct {
	itemRepo requisitionout.ItemRepository
}

func NewActivateItemUsecase(
	itemRepo requisitionout.ItemRepository,
) *ActivateItemUsecase {
	return &ActivateItemUsecase{
		itemRepo: itemRepo,
	}
}

func (uc *ActivateItemUsecase) Execute(
	ctx context.Context,
	id string,
) (*domain.Item, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	item, err := uc.itemRepo.GetItemByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := item.Activate(); err != nil {
		return nil, err
	}

	if err := uc.itemRepo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

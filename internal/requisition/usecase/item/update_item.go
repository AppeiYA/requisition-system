package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

type UpdateItemUsecase struct {
	itemRepo requisitionout.ItemRepository
}

func NewUpdateItemUsecase(
	itemRepo requisitionout.ItemRepository,
) *UpdateItemUsecase {
	return &UpdateItemUsecase{
		itemRepo: itemRepo,
	}
}

func (uc *UpdateItemUsecase) Execute(
	ctx context.Context,
	cmd itemin.UpdateItemCommand,
) (*domain.Item, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	item, err := uc.itemRepo.GetItemByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	if err := item.Update(cmd.Name, cmd.Description, cmd.CategoryID, cmd.UnitCost); err != nil {
		return nil, err
	}

	if err := uc.itemRepo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

type CreateItemUsecase struct {
	itemRepo    requisitionout.ItemRepository
	idGenerator requisitionout.IDGenerator
}

func NewCreateItemUsecase(
	itemRepo requisitionout.ItemRepository,
	idGenerator requisitionout.IDGenerator,
) *CreateItemUsecase {
	return &CreateItemUsecase{
		itemRepo:    itemRepo,
		idGenerator: idGenerator,
	}
}

func (uc *CreateItemUsecase) Execute(
	ctx context.Context,
	cmd itemin.CreateItemCommand,
) (*domain.Item, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	name, err := domain.NewItemName(cmd.Name)
	if err != nil {
		return nil, err
	}

	description, err := domain.NewItemDescription(cmd.Description)
	if err != nil {
		return nil, err
	}

	itemID, err := uc.idGenerator.GenerateID(domain.ItemPrefixID)
	if err != nil {
		return nil, err
	}

	item, err := domain.NewItem(itemID, name, description, cmd.CategoryID, cmd.UnitCost)
	if err != nil {
		return nil, err
	}

	if err := uc.itemRepo.SaveItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

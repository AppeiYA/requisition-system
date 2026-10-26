package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

type CreateItemCategoryUsecase struct {
	itemCategoryRepo requisitionout.ItemCategoryRepository
	idGenerator      requisitionout.IDGenerator
}

func NewCreateItemCategoryUsecase(
	itemCategoryRepo requisitionout.ItemCategoryRepository,
	idGenerator requisitionout.IDGenerator,
) *CreateItemCategoryUsecase {
	return &CreateItemCategoryUsecase{
		itemCategoryRepo: itemCategoryRepo,
		idGenerator:      idGenerator,
	}
}

func (uc *CreateItemCategoryUsecase) Execute(
	ctx context.Context,
	cmd requisitionin.CreateItemCategoryCommand,
) (*domain.ItemCategory, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	name, err := domain.NewCategoryName(cmd.Name)
	if err != nil {
		return nil, err
	}

	description, err := domain.NewCategoryDescription(cmd.Description)
	if err != nil {
		return nil, err
	}

	categoryID, err := uc.idGenerator.GenerateID(domain.CategoryPrefixID)
	if err != nil {
		return nil, err
	}

	category, err := domain.NewItemCategory(categoryID, name, description)
	if err != nil {
		return nil, err
	}

	if err := uc.itemCategoryRepo.SaveItemCategory(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

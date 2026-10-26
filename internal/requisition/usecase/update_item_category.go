package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

type UpdateItemCategoryUsecase struct {
	itemCategoryRepo requisitionout.ItemCategoryRepository
}

func NewUpdateItemCategoryUsecase(
	itemCategoryRepo requisitionout.ItemCategoryRepository,
) *UpdateItemCategoryUsecase {
	return &UpdateItemCategoryUsecase{
		itemCategoryRepo: itemCategoryRepo,
	}
}

func (uc *UpdateItemCategoryUsecase) Execute(
	ctx context.Context,
	cmd requisitionin.UpdateItemCategoryCommand,
) (*domain.ItemCategory, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	category, err := uc.itemCategoryRepo.GetItemCategoryByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	if err := category.Update(cmd.Name, cmd.Description); err != nil {
		return nil, err
	}

	if err := uc.itemCategoryRepo.UpdateItemCategory(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

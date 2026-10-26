package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

type ActivateItemCategoryUsecase struct {
	itemCategoryRepo requisitionout.ItemCategoryRepository
}

func NewActivateItemCategoryUsecase(
	itemCategoryRepo requisitionout.ItemCategoryRepository,
) *ActivateItemCategoryUsecase {
	return &ActivateItemCategoryUsecase{
		itemCategoryRepo: itemCategoryRepo,
	}
}

func (uc *ActivateItemCategoryUsecase) Execute(
	ctx context.Context,
	id string,
) (*domain.ItemCategory, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	category, err := uc.itemCategoryRepo.GetItemCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	category.Activate()

	if err := uc.itemCategoryRepo.UpdateItemCategory(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

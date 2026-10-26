package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

type GetItemCategoryUsecase struct {
	itemCategoryRepo requisitionout.ItemCategoryRepository
}

func NewGetItemCategoryUsecase(
	itemCategoryRepo requisitionout.ItemCategoryRepository,
) *GetItemCategoryUsecase {
	return &GetItemCategoryUsecase{
		itemCategoryRepo: itemCategoryRepo,
	}
}

func (uc *GetItemCategoryUsecase) Execute(
	ctx context.Context,
	id string,
) (*domain.ItemCategory, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return uc.itemCategoryRepo.GetItemCategoryByID(ctx, id)
}

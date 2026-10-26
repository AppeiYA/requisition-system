package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	categoryin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/category"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

type ListItemCategoriesUsecase struct {
	itemCategoryRepo requisitionout.ItemCategoryRepository
}

func NewListItemCategoriesUsecase(
	itemCategoryRepo requisitionout.ItemCategoryRepository,
) *ListItemCategoriesUsecase {
	return &ListItemCategoriesUsecase{
		itemCategoryRepo: itemCategoryRepo,
	}
}

func (uc *ListItemCategoriesUsecase) Execute(
	ctx context.Context,
	query categoryin.ListItemCategoriesQuery,
) ([]*domain.ItemCategory, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return uc.itemCategoryRepo.ListItemCategories(ctx, query.Limit, query.Offset)
}

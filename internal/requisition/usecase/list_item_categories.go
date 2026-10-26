package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
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
	query requisitionin.ListItemCategoriesQuery,
) ([]*domain.ItemCategory, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return uc.itemCategoryRepo.ListItemCategories(ctx, query.Limit, query.Offset)
}

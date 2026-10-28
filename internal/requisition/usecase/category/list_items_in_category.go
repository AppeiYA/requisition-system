package usecase

import (
	"context"
	"strings"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

type ListItemsInCategoryUsecase struct {
	categoryRepo requisitionout.ItemCategoryRepository
}

func NewListItemsInCategoryUsecase(
	categoryRepo requisitionout.ItemCategoryRepository,
) *ListItemsInCategoryUsecase {
	return &ListItemsInCategoryUsecase{
		categoryRepo: categoryRepo,
	}
}

func (uc *ListItemsInCategoryUsecase) Execute(ctx context.Context, categoryID string) ([]*domain.Item, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if categoryID == "" || !strings.HasPrefix(categoryID, domain.CategoryPrefixID) {
		return nil, domain.ErrInvalidCategoryID
	}
	return uc.categoryRepo.ListItemsInCategory(ctx, categoryID)
}
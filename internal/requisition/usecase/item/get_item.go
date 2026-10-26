package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

type GetItemUsecase struct {
	itemRepo requisitionout.ItemRepository
}

func NewGetItemUsecase(
	itemRepo requisitionout.ItemRepository,
) *GetItemUsecase {
	return &GetItemUsecase{
		itemRepo: itemRepo,
	}
}

func (uc *GetItemUsecase) Execute(
	ctx context.Context,
	id string,
) (*domain.Item, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return uc.itemRepo.GetItemByID(ctx, id)
}

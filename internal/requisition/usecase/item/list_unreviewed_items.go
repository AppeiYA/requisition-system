package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

type ListUnreviewedItemsUsecase struct {
	requestedItemRepo requisitionout.RequestedItemRepository
}

func NewListUnreviewedItemsUsecase(
	requestedItemRepo requisitionout.RequestedItemRepository,
) *ListUnreviewedItemsUsecase {
	return &ListUnreviewedItemsUsecase{
		requestedItemRepo: requestedItemRepo,
	}
}

func (uc *ListUnreviewedItemsUsecase) Execute(
	ctx context.Context,
	query itemin.ListItemsQuery,
) ([]*domain.RequestedItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	limit, offset := shareddomain.Clamp(query.Limit, query.Offset)
	return uc.requestedItemRepo.ListUnreviewedRequestedItems(ctx, limit, offset)
}

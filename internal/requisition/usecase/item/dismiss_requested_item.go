package usecase

import (
	"context"
	"time"

	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var ErrCannotDismissRequestedItem = customerrors.NewBadRequestError("requested item cannot be dismissed")

type DismissRequestedItemUsecase struct {
	requestedItemRepo requisitionout.RequestedItemRepository
	userLookup        requisitionout.UserLookup
	txManager         requisitionout.TransactionManager
}

func NewDismissRequestedItemUsecase(
	requestedItemRepo requisitionout.RequestedItemRepository,
	userLookup requisitionout.UserLookup,
	txManager requisitionout.TransactionManager,
) *DismissRequestedItemUsecase {
	return &DismissRequestedItemUsecase{
		requestedItemRepo: requestedItemRepo,
		userLookup:        userLookup,
		txManager:         txManager,
	}
}

func (uc *DismissRequestedItemUsecase) Execute(ctx context.Context, req itemin.DismissRequestedItemDTO) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if uc.txManager == nil {
		return ErrTransactionManagerNotProvided
	}

	return uc.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		user, err := uc.userLookup.GetUser(txCtx, req.ReviewerID)
		if err != nil {
			return err
		}
		if user == nil || user.Role != "admin" || user.Status != "active" {
			return ErrUserNotFound
		}

		requestedItem, err := uc.requestedItemRepo.GetRequestedItemByIDForUpdate(txCtx, req.RequestedItemID)
		if err != nil {
			return err
		}
		if err := requestedItem.MarkAsDismissed(req.ReviewerID, time.Now()); err != nil {
			if requestedItem.IsPromoted() || requestedItem.IsDismissed() {
				return ErrCannotDismissRequestedItem
			}
			return err
		}
		return uc.requestedItemRepo.UpdateRequestedItemStatus(txCtx, requestedItem)
	})
}

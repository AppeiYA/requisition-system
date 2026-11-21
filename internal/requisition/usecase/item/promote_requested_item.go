package usecase

import (
	"context"
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrUserNotFound                  = customerrors.NewBadRequestError("user is unauthorized to promote requested items")
	ErrTransactionManagerNotProvided = customerrors.NewInternalServerError("transaction manager not configured")
	ErrCannotPromoteRequestedItem    = customerrors.NewBadRequestError("requested item cannot be promoted")
)

type PromoteRequestedItemUsecase struct {
	requestedItemRepo requisitionout.RequestedItemRepository
	itemRepo          requisitionout.ItemRepository
	userLookup        requisitionout.UserLookup
	txManager         requisitionout.TransactionManager
	idGenerator       requisitionout.IDGenerator
}

func NewPromoteRequestedItemUsecase(
	requestedItemRepo requisitionout.RequestedItemRepository,
	itemRepo requisitionout.ItemRepository,
	userLookup requisitionout.UserLookup,
	txManager requisitionout.TransactionManager,
	idGenerator requisitionout.IDGenerator,
) *PromoteRequestedItemUsecase {
	return &PromoteRequestedItemUsecase{
		requestedItemRepo: requestedItemRepo,
		itemRepo:          itemRepo,
		userLookup:        userLookup,
		txManager:         txManager,
		idGenerator:       idGenerator,
	}
}

func (uc *PromoteRequestedItemUsecase) Execute(ctx context.Context, req itemin.PromoteRequestedItemDTO) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if uc.txManager == nil {
		return ErrTransactionManagerNotProvided
	}

	if err := uc.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// check if actor exists and has permission to promote items
		user, err := uc.userLookup.GetUser(txCtx, req.ReviewerID)
		if err != nil {
			return err
		}
		if user == nil {
			return ErrUserNotFound
		}
		if user.Role != "admin" || user.Status != "active" {
			return ErrUserNotFound
		}

		// get the requested item
		requestedItem, err := uc.requestedItemRepo.GetRequestedItemByIDForUpdate(txCtx, req.RequestedItemID)
		if err != nil {
			return err
		}
		if requestedItem == nil {
			return customerrors.NewNotFoundError("requested item not found")
		}
		if requestedItem.IsDismissed() || requestedItem.IsPromoted() {
			return ErrCannotPromoteRequestedItem
		}

		// promote the requested item to an actual item
		itemID, err := uc.idGenerator.GenerateID(domain.ItemPrefixID)
		if err != nil {
			return err
		}
		itemName, err := domain.NewItemName(requestedItem.Name().Value())
		if err != nil {
			return err
		}
		itemDescription, err := domain.NewItemDescription(requestedItem.Description().Value())
		if err != nil {
			return err
		}
		newItem, err := domain.NewItem(
			itemID,
			itemName,
			itemDescription,
			requestedItem.CategoryID(),
			req.UnitCost,
		)

		if err != nil {
			return err
		}

		if err := uc.itemRepo.SaveItem(txCtx, newItem); err != nil {
			return err
		}

		// update the requested item to mark it as promoted
		now := time.Now()
		if err := requestedItem.MarkAsPromoted(itemID, req.ReviewerID, now); err != nil {
			return err
		}
		if err := uc.requestedItemRepo.UpdateRequestedItemStatus(txCtx, requestedItem); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

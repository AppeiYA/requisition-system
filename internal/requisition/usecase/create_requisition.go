package usecase

import (
	"context"
	"strings"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrInactiveCategory = customerrors.NewBadRequestError("selected category is inactive")
	ErrItemNotInCategory = customerrors.NewBadRequestError("selected item does not belong to the selected category")
)

type CreateRequisitionUsecase struct {
	requisitionRepo requisitionout.RequisitionRepository
	categoryRepo    requisitionout.ItemCategoryRepository
	itemRepo        requisitionout.ItemRepository
	membership      requisitionout.MembershipLookup
	idGenerator     requisitionout.IDGenerator
}

func NewCreateRequisitionUsecase(
	requisitionRepo requisitionout.RequisitionRepository,
	categoryRepo requisitionout.ItemCategoryRepository,
	itemRepo requisitionout.ItemRepository,
	membership requisitionout.MembershipLookup,
	idGenerator requisitionout.IDGenerator,
) *CreateRequisitionUsecase {
	return &CreateRequisitionUsecase{
		requisitionRepo: requisitionRepo,
		categoryRepo:    categoryRepo,
		itemRepo:        itemRepo,
		membership:      membership,
		idGenerator:     idGenerator,
	}
}

func (uc *CreateRequisitionUsecase) Execute(ctx context.Context, cmd requisitionin.CreateRequisitionCommand) (*domain.Requisition, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if cmd.CategoryID == "" || !strings.HasPrefix(cmd.CategoryID, domain.CategoryPrefixID) {
		return nil, domain.ErrInvalidCategoryID
	}

	membership, err := uc.membership.GetMembershipByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	category, err := uc.categoryRepo.GetItemCategoryByID(ctx, cmd.CategoryID)
	if err != nil {
		return nil, err
	}
	if category.Status() != domain.CategoryStatusActive {
		return nil, ErrInactiveCategory
	}
	var itemName *domain.ItemName
	var itemDescription *domain.ItemDescription

	if cmd.ItemID != "" && strings.HasPrefix(cmd.ItemID, domain.ItemPrefixID) {
		item, err := uc.itemRepo.GetItemByID(ctx, cmd.ItemID)
		if err != nil {
			return nil, err
		}
		if item.CategoryID() != cmd.CategoryID {
			return nil, ErrItemNotInCategory
		}

		itemName = item.Name()
		itemDescription = item.Description()
	}else {
		itemName, err = domain.NewItemName(cmd.RequestedItemName)
		if err != nil {
			return nil, err
		}
		itemDescription, err = domain.NewItemDescription(cmd.RequestedItemDescription)
		if err != nil {
			return nil, err
		}
	}

	quantity, err := domain.NewRequisitionQuantity(cmd.Quantity)
	if err != nil {
		return nil, err
	}

	estimatedCost, err := shareddomain.NewNairaMoney(cmd.EstimatedCost)
	if err != nil {
		return nil, err
	}
	
	justification, err := domain.NewRequisitionJustification(cmd.Justification)
	if err != nil {
		return nil, err
	}
	priority, err := domain.NewRequisitionPriority(cmd.Priority)
	if err != nil {
		return nil, err
	}
	id, err := uc.idGenerator.GenerateID(domain.RequisitionPrefixID)
	if err != nil {
		return nil, err
	}

	requisition, err := domain.NewRequisition(
		id,
		cmd.UserID,
		membership.DepartmentID(),
		cmd.ItemID,
		cmd.CategoryID,
		itemName,
		itemDescription,
		quantity,
		estimatedCost,
		justification,
		priority,
		cmd.NeededBy,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.requisitionRepo.SaveRequisition(ctx, requisition); err != nil {
		return nil, err
	}
	return requisition, nil
}

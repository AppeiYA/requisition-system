package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrInactiveCategory   = customerrors.NewBadRequestError("selected category is inactive")
	ErrItemNotInCategory  = customerrors.NewBadRequestError("selected item does not belong to the selected category")
	ErrInactiveItem       = customerrors.NewBadRequestError("selected item is inactive")
	ErrInactiveMembership = customerrors.NewForbiddenError("department membership is not active")
	ErrDepartmentMismatch = customerrors.NewForbiddenError("requisition department does not match the authenticated user's department")
)

type CreateRequisitionUsecase struct {
	requisitionRepo requisitionout.RequisitionRepository
	categoryRepo    requisitionout.ItemCategoryRepository
	itemRepo        requisitionout.ItemRepository
	membership      requisitionout.MembershipLookup
	idGenerator     requisitionout.IDGenerator
	txManager       requisitionout.TransactionManager
}

func NewCreateRequisitionUsecase(
	requisitionRepo requisitionout.RequisitionRepository,
	categoryRepo requisitionout.ItemCategoryRepository,
	itemRepo requisitionout.ItemRepository,
	membership requisitionout.MembershipLookup,
	idGenerator requisitionout.IDGenerator,
	txManager requisitionout.TransactionManager,
) *CreateRequisitionUsecase {
	return &CreateRequisitionUsecase{
		requisitionRepo: requisitionRepo, categoryRepo: categoryRepo, itemRepo: itemRepo,
		membership: membership, idGenerator: idGenerator, txManager: txManager,
	}
}

func (uc *CreateRequisitionUsecase) Execute(ctx context.Context, cmd requisitionin.CreateRequisitionCommand) (*domain.Requisition, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	membership, err := uc.membership.GetMembershipByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}
	if membership.Status() != departmentdomain.MembershipStatusActive {
		return nil, ErrInactiveMembership
	}

	departmentID := cmd.DepartmentID
	if departmentID == "" {
		departmentID = membership.DepartmentID()
	}
	if !strings.HasPrefix(departmentID, domain.DepartmentPrefixID) {
		return nil, domain.ErrInvalidDepartmentID
	}
	if departmentID != membership.DepartmentID() {
		return nil, ErrDepartmentMismatch
	}

	lines := cmd.Lines
	if len(lines) == 0 {
		return nil, domain.ErrRequisitionNeedsItem
	}
	priority := cmd.Priority
	if priority == "" {
		priority = domain.PriorityMedium.String()
	} else if priority == domain.PriorityNormal.String() {
		priority = domain.PriorityMedium.String()
	}
	referenceCode := cmd.ReferenceCode
	if referenceCode == "" {
		referenceCode, err = uc.idGenerator.GenerateID(fmt.Sprintf("REQ-%d", time.Now().Year()))
		if err != nil {
			return nil, err
		}
	}
	requisitionID, err := uc.idGenerator.GenerateID(domain.RequisitionPrefixID)
	if err != nil {
		return nil, err
	}

	items := make([]*domain.RequisitionItem, 0, len(lines))
	requestedItems := make([]*domain.RequestedItem, 0)
	for lineNumber, line := range lines {
		item, requestedItem, err := uc.buildLine(ctx, requisitionID, cmd.UserID, lineNumber+1, line)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		if requestedItem != nil {
			requestedItems = append(requestedItems, requestedItem)
		}
	}

	justification, err := domain.NewRequisitionJustification(cmd.Justification)
	if err != nil {
		return nil, err
	}
	priorityValue, err := domain.NewRequisitionPriority(priority)
	if err != nil {
		return nil, err
	}
	requisition, err := domain.NewSubmittedRequisitionWithRequestedItems(
		requisitionID, referenceCode, cmd.UserID, departmentID,
		priorityValue, cmd.NeededBy, justification, items, requestedItems,
	)
	if err != nil {
		return nil, err
	}
	historyID, err := uc.idGenerator.GenerateID(domain.StatusHistoryPrefixID)
	if err != nil {
		return nil, err
	}
	history, err := requisition.InitialStatusHistory(
		historyID, cmd.UserID, domain.ActorRoleStaff, time.Now(),
	)
	if err != nil {
		return nil, err
	}
	if err := requisition.RecordInitialStatusHistory(history); err != nil {
		return nil, err
	}

	if uc.txManager == nil {
		return nil, customerrors.NewInternalServerError("requisition transaction manager is not configured")
	}
	if err := uc.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		return uc.requisitionRepo.SaveRequisition(txCtx, requisition)
	}); err != nil {
		return nil, err
	}
	return requisition, nil
}

func (uc *CreateRequisitionUsecase) buildLine(
	ctx context.Context,
	requisitionID string,
	createdBy string,
	lineNumber int,
	cmd requisitionin.CreateRequisitionLine,
) (*domain.RequisitionItem, *domain.RequestedItem, error) {
	return buildRequisitionLine(
		ctx, uc.categoryRepo, uc.itemRepo, uc.idGenerator,
		requisitionID, createdBy, lineNumber, cmd,
	)
}

func buildRequisitionLine(
	ctx context.Context,
	categoryRepo requisitionout.ItemCategoryRepository,
	itemRepo requisitionout.ItemRepository,
	idGenerator requisitionout.IDGenerator,
	requisitionID string,
	createdBy string,
	lineNumber int,
	cmd requisitionin.CreateRequisitionLine,
) (*domain.RequisitionItem, *domain.RequestedItem, error) {
	hasCatalogueItem := cmd.ItemID != ""
	hasRequestedItem := cmd.RequestedItem != nil
	if hasCatalogueItem == hasRequestedItem {
		return nil, nil, domain.ErrInvalidItemSelection
	}

	itemID := cmd.ItemID
	requestedItemID := ""
	var name *domain.ItemName
	var requestedItem *domain.RequestedItem

	if hasCatalogueItem {
		if !strings.HasPrefix(itemID, domain.ItemPrefixID) {
			return nil, nil, domain.ErrInvalidItemID
		}
		item, err := itemRepo.GetItemByID(ctx, itemID)
		if err != nil {
			return nil, nil, err
		}
		if item.Status() != domain.ItemStatusActive {
			return nil, nil, ErrInactiveItem
		}
		category, err := categoryRepo.GetItemCategoryByID(ctx, item.CategoryID())
		if err != nil {
			return nil, nil, err
		}
		if category.Status() != domain.CategoryStatusActive {
			return nil, nil, ErrInactiveCategory
		}
		name = item.Name()
	} else {
		requested := cmd.RequestedItem
		if !strings.HasPrefix(requested.CategoryID, domain.CategoryPrefixID) {
			return nil, nil, domain.ErrInvalidCategoryID
		}
		category, err := categoryRepo.GetItemCategoryByID(ctx, requested.CategoryID)
		if err != nil {
			return nil, nil, err
		}
		if category.Status() != domain.CategoryStatusActive {
			return nil, nil, ErrInactiveCategory
		}
		name, err = domain.NewItemName(requested.Name)
		if err != nil {
			return nil, nil, err
		}
		description, err := domain.NewItemDescription(requested.Description)
		if err != nil {
			return nil, nil, err
		}
		requestedItemID, err = idGenerator.GenerateID(domain.RequestedItemPrefixID)
		if err != nil {
			return nil, nil, err
		}
		requestedItem, err = domain.NewRequestedItem(
			requestedItemID, name, requested.CategoryID, description, createdBy,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	quantity, err := domain.NewRequisitionQuantity(cmd.Quantity)
	if err != nil {
		return nil, nil, err
	}
	amount, err := shareddomain.NewNairaMoney(cmd.EstimatedUnitAmount)
	if err != nil {
		return nil, nil, err
	}
	lineID, err := idGenerator.GenerateID(domain.RequisitionItemPrefixID)
	if err != nil {
		return nil, nil, err
	}
	line, err := domain.NewRequisitionItem(
		lineID, requisitionID, lineNumber, itemID, requestedItemID,
		name, quantity, amount, cmd.Notes,
	)
	return line, requestedItem, err
}

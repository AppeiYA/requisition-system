package usecase

import (
	"context"
	"strings"
	"time"

	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

type UpdateRequisitionUsecase struct {
	requisitionRepo requisitionout.RequisitionEditRepository
	categoryRepo    requisitionout.ItemCategoryRepository
	itemRepo        requisitionout.ItemRepository
	membership      requisitionout.MembershipLookup
	idGenerator     requisitionout.IDGenerator
	txManager       requisitionout.TransactionManager
}

func NewUpdateRequisitionUsecase(
	requisitionRepo requisitionout.RequisitionEditRepository,
	categoryRepo requisitionout.ItemCategoryRepository,
	itemRepo requisitionout.ItemRepository,
	membership requisitionout.MembershipLookup,
	idGenerator requisitionout.IDGenerator,
	txManager requisitionout.TransactionManager,
) *UpdateRequisitionUsecase {
	return &UpdateRequisitionUsecase{
		requisitionRepo: requisitionRepo, categoryRepo: categoryRepo, itemRepo: itemRepo,
		membership: membership, idGenerator: idGenerator, txManager: txManager,
	}
}

func (uc *UpdateRequisitionUsecase) Execute(
	ctx context.Context,
	cmd requisitionin.UpdateRequisitionCommand,
) (*domain.Requisition, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !strings.HasPrefix(cmd.RequisitionID, domain.RequisitionPrefixID) {
		return nil, domain.ErrInvalidRequisitionID
	}
	if cmd.UserID == "" {
		return nil, domain.ErrRequestedIDCannotBeEmpty
	}
	if len(cmd.Lines) == 0 {
		return nil, domain.ErrRequisitionNeedsItem
	}
	if uc.txManager == nil {
		return nil, customerrors.NewInternalServerError("requisition transaction manager is not configured")
	}

	var updated *domain.Requisition
	err := uc.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		requisition, err := uc.requisitionRepo.GetRequisitionByIDForUpdate(txCtx, cmd.RequisitionID)
		if err != nil {
			return err
		}
		if requisition.RequestedBy() != cmd.UserID {
			return ErrRequisitionAccessDenied
		}
		if !requisition.IsEditable() {
			return domain.ErrCannotUpdateRequisition
		}

		membership, err := uc.membership.GetMembershipByUserID(txCtx, cmd.UserID)
		if err != nil {
			return err
		}
		if membership.Status() != departmentdomain.MembershipStatusActive {
			return ErrInactiveMembership
		}
		departmentID := cmd.DepartmentID
		if departmentID == "" {
			departmentID = requisition.DepartmentID()
		}
		if departmentID != membership.DepartmentID() {
			return ErrDepartmentMismatch
		}

		priority := cmd.Priority
		if priority == "" {
			priority = requisition.Priority().String()
		}
		justification, err := domain.NewRequisitionJustification(cmd.Justification)
		if err != nil {
			return err
		}
		priorityValue, err := domain.NewRequisitionPriority(priority)
		if err != nil {
			return err
		}
		if isDateNotFuture(cmd.NeededBy) {
			return domain.ErrInvalidNeededByDate
		}

		existingLines := make(map[string]*domain.RequisitionItem, len(requisition.Items()))
		for _, item := range requisition.Items() {
			existingLines[item.ID()] = item
		}
		existingRequestedItems := make(map[string]*domain.RequestedItem, len(requisition.RequestedItems()))
		for _, requestedItem := range requisition.RequestedItems() {
			existingRequestedItems[requestedItem.ID()] = requestedItem
		}

		items := make([]*domain.RequisitionItem, 0, len(cmd.Lines))
		requestedItems := make([]*domain.RequestedItem, 0)
		seenLineIDs := make(map[string]struct{}, len(cmd.Lines))
		for lineNumber, line := range cmd.Lines {
			var identity *lineIdentity
			if line.RequisitionItemID != "" {
				if !strings.HasPrefix(line.RequisitionItemID, domain.RequisitionItemPrefixID) {
					return domain.ErrInvalidRequisitionItemID
				}
				if _, duplicate := seenLineIDs[line.RequisitionItemID]; duplicate {
					return domain.ErrDuplicateRequisitionItemID
				}
				existingLine, exists := existingLines[line.RequisitionItemID]
				if !exists {
					return domain.ErrRequisitionItemNotOwned
				}
				seenLineIDs[line.RequisitionItemID] = struct{}{}
				identity = &lineIdentity{line: existingLine}
				if existingLine.RequestedItemID() != "" {
					existingRequestedItem, exists := existingRequestedItems[existingLine.RequestedItemID()]
					if !exists {
						return domain.ErrInvalidRequestedItemID
					}
					identity.requestedItem = existingRequestedItem
				}
			}

			item, requestedItem, err := buildRequisitionLine(
				txCtx, uc.categoryRepo, uc.itemRepo, uc.idGenerator,
				requisition.ID(), cmd.UserID, lineNumber+1,
				requisitionin.CreateRequisitionLine{
					ItemID:              line.ItemID,
					RequestedItem:       line.RequestedItem,
					Quantity:            line.Quantity,
					EstimatedUnitAmount: line.EstimatedUnitAmount,
					Notes:               line.Notes,
				},
				identity,
			)
			if err != nil {
				return err
			}
			items = append(items, item)
			if requestedItem != nil {
				requestedItems = append(requestedItems, requestedItem)
			}
		}
		if err := requisition.ReplaceItems(items, requestedItems); err != nil {
			return err
		}
		if err := requisition.UpdateRequisition(&domain.UpdateRequisitionInput{
			Justification: justification, Priority: &priorityValue, NeededBy: &cmd.NeededBy,
		}); err != nil {
			return err
		}
		// The repository requires the transaction context and the header lock
		// acquired by GetRequisitionByIDForUpdate above.
		if err := uc.requisitionRepo.UpdateRequisitionAggregate(txCtx, requisition); err != nil {
			return err
		}
		updated = requisition
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func isDateNotFuture(value time.Time) bool {
	today := time.Now().UTC()
	date := value.UTC()
	return !date.After(time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC))
}

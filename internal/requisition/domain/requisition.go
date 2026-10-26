package domain

import (
	"strings"
	"time"

	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var RequisitionPrefixID = "req"

var (
	ErrInvalidRequisitionID     = customerrors.NewBadRequestError("invalid requisition ID")
	ErrRequestedIDCannotBeEmpty = customerrors.NewBadRequestError("requestedBy cannot be empty")
	ErrNilFields                = func(fieldName string) error {
		return customerrors.NewBadRequestError(fieldName + " cannot be nil")
	}
	ErrInvalidNeededByDate = customerrors.NewBadRequestError("neededBy date cannot be in the past")
)

type Requisition struct {
	id                       string
	requestedBy              string
	itemID                   string
	categoryID               string
	requestedItemName        *ItemName
	requestedItemDescription *ItemDescription
	quantity                 *RequisitionQuantity
	estimatedCost            *shareddomain.Money
	justification            *RequisitionJustification
	priority                 RequisitionPriority
	neededBy                 time.Time
	status                   RequisitionStatus
	createdAt                time.Time
	updatedAt                time.Time
}

func NewRequisition(id string, requestedBy string, itemID string, categoryID string, requestedItemName *ItemName, requestedItemDescription *ItemDescription, quantity *RequisitionQuantity, estimatedCost *shareddomain.Money, justification *RequisitionJustification, priority RequisitionPriority, neededBy time.Time) (*Requisition, error) {
	if id == "" || !strings.HasPrefix(id, RequisitionPrefixID) {
		return nil, ErrInvalidRequisitionID
	}

	if requestedBy == "" {
		return nil, ErrRequestedIDCannotBeEmpty
	}

	if categoryID != "" && !strings.HasPrefix(categoryID, CategoryPrefixID) {
		return nil, ErrInvalidCategoryID
	}

	if requestedItemName == nil {
		return nil, ErrNilFields("requestedItemName")
	}

	if requestedItemDescription == nil {
		return nil, ErrNilFields("requestedItemDescription")
	}

	if justification == nil {
		return nil, ErrNilFields("justification")
	}

	if quantity == nil {
		return nil, ErrNilFields("quantity")
	}

	if estimatedCost == nil {
		return nil, ErrNilFields("estimatedCost")
	}

	if neededBy.Before(time.Now()) {
		return nil, ErrInvalidNeededByDate
	}

	if itemID != "" && !strings.HasPrefix(itemID, ItemPrefixID) {
		return nil, ErrInvalidItemID
	}
	if !priority.IsValid() {
		return nil, ErrInvalidRequisitionPriority
	}
	now := time.Now()

	return &Requisition{
		id:                       id,
		requestedBy:              requestedBy,
		categoryID:               categoryID,
		itemID:                   itemID,
		requestedItemName:        requestedItemName,
		requestedItemDescription: requestedItemDescription,
		quantity:                 quantity,
		estimatedCost:            estimatedCost,
		justification:            justification,
		priority:                 priority,
		neededBy:                 neededBy,
		status:                   RequisitionStatusPending,
		createdAt:                now,
		updatedAt:                now,
	}, nil
}

func ReconstituteRequisition(
	id string,
	requestedBy string,
	itemID string,
	categoryID string,
	requestedItemName string,
	requestedItemDescription string,
	quantity int,
	estimatedCost *shareddomain.Money,
	justification string,
	priority string,
	neededBy time.Time,
	status string,
	createdAt time.Time,
	updatedAt time.Time,
) (*Requisition, error) {
	if id == "" || !strings.HasPrefix(id, RequisitionPrefixID) {
		return nil, ErrInvalidRequisitionID
	}

	if requestedBy == "" {
		return nil, ErrRequestedIDCannotBeEmpty
	}

	if categoryID != "" && !strings.HasPrefix(categoryID, CategoryPrefixID) {
		return nil, ErrInvalidCategoryID
	}

	if itemID != "" && !strings.HasPrefix(itemID, ItemPrefixID) {
		return nil, ErrInvalidItemID
	}

	requestedItemNameObj, err := NewItemName(requestedItemName)
	if err != nil {
		return nil, err
	}

	requestedItemDescriptionObj, err := NewItemDescription(requestedItemDescription)
	if err != nil {
		return nil, err
	}

	quantityObj, err := NewRequisitionQuantity(quantity)
	if err != nil {
		return nil, err
	}

	justificationObj, err := NewRequisitionJustification(justification)
	if err != nil {
		return nil, err
	}

	requisitionPriority, err := NewRequisitionPriority(priority)
	if err != nil {
		return nil, err
	}

	requisitionStatus, err := NewRequisitionStatus(status)
	if err != nil {
		return nil, err
	}

	return &Requisition{
		id:                       id,
		requestedBy:              requestedBy,
		itemID:                   itemID,
		categoryID:               categoryID,
		requestedItemName:        requestedItemNameObj,
		requestedItemDescription: requestedItemDescriptionObj,
		quantity:                 quantityObj,
		estimatedCost:            estimatedCost,
		justification:            justificationObj,
		priority:                 requisitionPriority,
		neededBy:                 neededBy,
		status:                   requisitionStatus,
		createdAt:                createdAt,
		updatedAt:                updatedAt,
	}, nil
}

func (r *Requisition) ID() string {
	return r.id
}

func (r *Requisition) RequestedBy() string {
	return r.requestedBy
}

func (r *Requisition) ItemID() string {
	return r.itemID
}

func (r *Requisition) CategoryID() string {
	return r.categoryID
}

func (r *Requisition) RequestedItemName() *ItemName {
	return r.requestedItemName
}

func (r *Requisition) RequestedItemDescription() *ItemDescription {
	return r.requestedItemDescription
}

func (r *Requisition) Quantity() *RequisitionQuantity {
	return r.quantity
}

func (r *Requisition) EstimatedCost() *shareddomain.Money {
	return r.estimatedCost
}

func (r *Requisition) Justification() *RequisitionJustification {
	return r.justification
}

func (r *Requisition) Priority() RequisitionPriority {
	return r.priority
}

func (r *Requisition) NeededBy() time.Time {
	return r.neededBy
}

func (r *Requisition) Status() RequisitionStatus {
	return r.status
}

func (r *Requisition) CreatedAt() time.Time {
	return r.createdAt
}

func (r *Requisition) UpdatedAt() time.Time {
	return r.updatedAt
}

func (r *Requisition) UpdateStatus(newStatus RequisitionStatus) bool {
	if !r.status.CanTransitionTo(newStatus) {
		return false
	}
	r.status = newStatus
	r.updatedAt = time.Now()
	return true
}

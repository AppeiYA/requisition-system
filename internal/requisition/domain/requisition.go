package domain

import (
	"strings"
	"time"

	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	RequisitionPrefixID = "req"
	DepartmentPrefixID  = "dpt"
)

var (
	ErrInvalidDepartmentID       = customerrors.NewBadRequestError("invalid department ID")
	ErrNilFields                 = func(fieldName string) error {
		return customerrors.NewBadRequestError(fieldName + " cannot be nil")
	}
	ErrCannotTransitionToStatus = func(currentStatus, newStatus string) error {
		return customerrors.NewBadRequestError("cannot transition from " + currentStatus + " to " + newStatus)
	}
	ErrInvalidReferenceCode   = customerrors.NewBadRequestError("invalid requisition reference code")
	ErrRequisitionNeedsItem   = customerrors.NewBadRequestError("requisition must contain at least one item")
	ErrInvalidRequisitionID    = customerrors.NewBadRequestError("invalid requisition ID")
	ErrInvalidNeededByDate       = customerrors.NewBadRequestError("needed by date cannot be in the past")
	ErrDepartmentIDCannotBeEmpty = customerrors.NewBadRequestError("department ID cannot be empty")
	ErrRequestedIDCannotBeEmpty = customerrors.NewBadRequestError("requested by ID cannot be empty")
	ErrUpdateInputCannotBeNil = customerrors.NewBadRequestError("update input cannot be nil")
	ErrCannotUpdateRequisition = customerrors.NewBadRequestError("cannot update requisition")
)

type Requisition struct {
	id                       string
	referenceCode            string
	requestedBy              string
	departmentID             string
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
	items                    []*RequisitionItem
	requestedItems           []*RequestedItem
	statusHistory            []*RequisitionStatusHistory
}

type UpdateRequisitionInput struct {
	Quantity      *RequisitionQuantity
	EstimatedCost *shareddomain.Money
	Justification *RequisitionJustification
	Priority      *RequisitionPriority
	NeededBy      *time.Time
}

func NewRequisition(id string, requestedBy string, departmentID string, itemID string, categoryID string, requestedItemName *ItemName, requestedItemDescription *ItemDescription, quantity *RequisitionQuantity, estimatedCost *shareddomain.Money, justification *RequisitionJustification, priority RequisitionPriority, neededBy time.Time) (*Requisition, error) {
	if id == "" || !strings.HasPrefix(id, RequisitionPrefixID) {
		return nil, ErrInvalidRequisitionID
	}

	if requestedBy == "" {
		return nil, ErrRequestedIDCannotBeEmpty
	}

	if departmentID == "" {
		return nil, ErrDepartmentIDCannotBeEmpty
	}
	if !strings.HasPrefix(departmentID, DepartmentPrefixID) {
		return nil, ErrInvalidDepartmentID
	}

	if categoryID == "" || !strings.HasPrefix(categoryID, CategoryPrefixID) {
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

	if isDateBeforeToday(neededBy) {
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
		departmentID:             departmentID,
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
	departmentID string,
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

	if departmentID == "" {
		return nil, ErrDepartmentIDCannotBeEmpty
	}
	if !strings.HasPrefix(departmentID, DepartmentPrefixID) {
		return nil, ErrInvalidDepartmentID
	}

	if categoryID == "" || !strings.HasPrefix(categoryID, CategoryPrefixID) {
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
		departmentID:             departmentID,
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

func (r *Requisition) DepartmentID() string {
	return r.departmentID
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

func (r *Requisition) UpdateRequisition(input *UpdateRequisitionInput) error {
	if input == nil {
		return ErrUpdateInputCannotBeNil
	}

	if r.status != RequisitionStatusPending {
		return ErrCannotUpdateRequisition
	}

	if input.Quantity != nil {
		r.quantity = input.Quantity
	}

	if input.EstimatedCost != nil {
		r.estimatedCost = input.EstimatedCost
	}

	if input.Justification != nil {
		r.justification = input.Justification
	}

	if input.Priority != nil {
		if !input.Priority.IsValid() {
			return ErrInvalidRequisitionPriority
		}
		r.priority = *input.Priority
	}

	if input.NeededBy != nil {
		if isDateBeforeToday(*input.NeededBy) {
			return ErrInvalidNeededByDate
		}
		r.neededBy = *input.NeededBy
	}

	r.updatedAt = time.Now()
	return nil
}

func isDateBeforeToday(value time.Time) bool {
	valueDate := value.UTC()
	today := time.Now().UTC()
	if valueDate.Year() != today.Year() {
		return valueDate.Year() < today.Year()
	}
	if valueDate.Month() != today.Month() {
		return valueDate.Month() < today.Month()
	}
	return valueDate.Day() < today.Day()
}

func (r *Requisition) SetApproved() error {
	if !r.status.CanTransitionTo(RequisitionStatusApproved) {
		return ErrCannotTransitionToStatus(r.status.String(), RequisitionStatusApproved.String())
	}
	r.status = RequisitionStatusApproved
	r.updatedAt = time.Now()
	return nil
}

func (r *Requisition) SetRejected() error {
	if !r.status.CanTransitionTo(RequisitionStatusRejected) {
		return ErrCannotTransitionToStatus(r.status.String(), RequisitionStatusRejected.String())
	}
	r.status = RequisitionStatusRejected
	r.updatedAt = time.Now()
	return nil
}

func (r *Requisition) SetInFulfillment() error {
	if !r.status.CanTransitionTo(RequisitionStatusInFulfillment) {
		return ErrCannotTransitionToStatus(r.status.String(), RequisitionStatusInFulfillment.String())
	}
	r.status = RequisitionStatusInFulfillment
	r.updatedAt = time.Now()
	return nil
}

func (r *Requisition) SetCompleted() error {
	if !r.status.CanTransitionTo(RequisitionStatusCompleted) {
		return ErrCannotTransitionToStatus(r.status.String(), RequisitionStatusCompleted.String())
	}
	r.status = RequisitionStatusCompleted
	r.updatedAt = time.Now()
	return nil
}

func (r *Requisition) SetCancelled() error {
	if !r.status.CanTransitionTo(RequisitionStatusCancelled) {
		return ErrCannotTransitionToStatus(r.status.String(), RequisitionStatusCancelled.String())
	}
	r.status = RequisitionStatusCancelled
	r.updatedAt = time.Now()
	return nil
}

func (r *Requisition) UpdateStatus(newStatus RequisitionStatus) bool {
	if !r.status.CanTransitionTo(newStatus) {
		return false
	}
	r.status = newStatus
	r.updatedAt = time.Now()
	return true
}

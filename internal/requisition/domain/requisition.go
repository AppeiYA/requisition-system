package domain

import (
	"time"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	RequisitionPrefixID = "req"
	DepartmentPrefixID  = "dpt"
)

var (
	ErrInvalidDepartmentID = customerrors.NewBadRequestError("invalid department ID")
	ErrNilFields           = func(fieldName string) error {
		return customerrors.NewBadRequestError(fieldName + " cannot be nil")
	}
	ErrCannotTransitionToStatus = func(currentStatus, newStatus string) error {
		return customerrors.NewBadRequestError("cannot transition from " + currentStatus + " to " + newStatus)
	}
	ErrInvalidReferenceCode      = customerrors.NewBadRequestError("invalid requisition reference code")
	ErrRequisitionNeedsItem      = customerrors.NewBadRequestError("requisition must contain at least one item")
	ErrInvalidRequisitionID      = customerrors.NewBadRequestError("invalid requisition ID")
	ErrInvalidNeededByDate       = customerrors.NewBadRequestError("needed by date cannot be in the past")
	ErrDepartmentIDCannotBeEmpty = customerrors.NewBadRequestError("department ID cannot be empty")
	ErrRequestedIDCannotBeEmpty  = customerrors.NewBadRequestError("requested by ID cannot be empty")
	ErrUpdateInputCannotBeNil    = customerrors.NewBadRequestError("update input cannot be nil")
	ErrCannotUpdateRequisition   = customerrors.NewBadRequestError("cannot update requisition")
)

type Requisition struct {
	id             string
	referenceCode  string
	requestedBy    string
	departmentID   string
	justification  *RequisitionJustification
	priority       RequisitionPriority
	neededBy       time.Time
	status         RequisitionStatus
	createdAt      time.Time
	updatedAt      time.Time
	items          []*RequisitionItem
	requestedItems []*RequestedItem
	statusHistory  []*RequisitionStatusHistory
}

type UpdateRequisitionInput struct {
	Justification *RequisitionJustification
	Priority      *RequisitionPriority
	NeededBy      *time.Time
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
		if isDateNotFuture(*input.NeededBy) {
			return ErrInvalidNeededByDate
		}
		r.neededBy = *input.NeededBy
	}

	r.updatedAt = time.Now()
	return nil
}

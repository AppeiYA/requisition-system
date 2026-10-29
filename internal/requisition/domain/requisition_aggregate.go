package domain

import (
	"strings"
	"time"

	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

// NewSubmittedRequisition creates the phase-one aggregate. Persistence and
// reference-code generation remain outside the domain.
func NewSubmittedRequisition(
	id string,
	referenceCode string,
	requestedBy string,
	departmentID string,
	priority RequisitionPriority,
	neededBy time.Time,
	justification *RequisitionJustification,
	items []*RequisitionItem,
) (*Requisition, error) {
	return NewSubmittedRequisitionWithRequestedItems(
		id, referenceCode, requestedBy, departmentID, priority, neededBy,
		justification, items, nil,
	)
}

func NewSubmittedRequisitionWithRequestedItems(
	id string,
	referenceCode string,
	requestedBy string,
	departmentID string,
	priority RequisitionPriority,
	neededBy time.Time,
	justification *RequisitionJustification,
	items []*RequisitionItem,
	requestedItems []*RequestedItem,
) (*Requisition, error) {
	if referenceCode == "" || len(referenceCode) > 30 {
		return nil, ErrInvalidReferenceCode
	}
	if len(items) == 0 {
		return nil, ErrRequisitionNeedsItem
	}
	if justification == nil {
		return nil, ErrNilFields("justification")
	}
	if err := validateRequisitionItems(id, items); err != nil {
		return nil, err
	}
	if err := validateRequestedItems(items, requestedItems); err != nil {
		return nil, err
	}

	requisition, err := newRequisitionHeader(id, requestedBy, departmentID, priority, neededBy, justification)
	if err != nil {
		return nil, err
	}
	requisition.referenceCode = referenceCode
	requisition.items = append([]*RequisitionItem(nil), items...)
	requisition.requestedItems = append([]*RequestedItem(nil), requestedItems...)
	return requisition, nil
}

func ReconstituteRequisitionAggregate(
	id string,
	referenceCode string,
	requestedBy string,
	departmentID string,
	priority string,
	neededBy time.Time,
	justification string,
	status string,
	createdAt time.Time,
	updatedAt time.Time,
	items []*RequisitionItem,
	history []*RequisitionStatusHistory,
) (*Requisition, error) {
	return ReconstituteRequisitionAggregateWithRequestedItems(
		id, referenceCode, requestedBy, departmentID, priority, neededBy,
		justification, status, createdAt, updatedAt, items, nil, history,
	)
}

func ReconstituteRequisitionAggregateWithRequestedItems(
	id string,
	referenceCode string,
	requestedBy string,
	departmentID string,
	priority string,
	neededBy time.Time,
	justification string,
	status string,
	createdAt time.Time,
	updatedAt time.Time,
	items []*RequisitionItem,
	requestedItems []*RequestedItem,
	history []*RequisitionStatusHistory,
) (*Requisition, error) {
	if referenceCode == "" || len(referenceCode) > 30 {
		return nil, ErrInvalidReferenceCode
	}
	if len(items) == 0 {
		return nil, ErrRequisitionNeedsItem
	}
	if err := validateRequisitionItems(id, items); err != nil {
		return nil, err
	}
	if err := validateRequestedItems(items, requestedItems); err != nil {
		return nil, err
	}
	priorityValue, err := NewRequisitionPriority(priority)
	if err != nil {
		return nil, err
	}
	statusValue, err := NewRequisitionStatus(status)
	if err != nil {
		return nil, err
	}
	justificationValue, err := NewRequisitionJustification(justification)
	if err != nil {
		return nil, err
	}
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
	requisition := &Requisition{
		id: id, referenceCode: referenceCode, requestedBy: requestedBy,
		departmentID: departmentID, priority: priorityValue, neededBy: neededBy,
		justification: justificationValue, status: statusValue,
		createdAt: createdAt, updatedAt: updatedAt,
		items:          append([]*RequisitionItem(nil), items...),
		requestedItems: append([]*RequestedItem(nil), requestedItems...),
		statusHistory:  append([]*RequisitionStatusHistory(nil), history...),
	}
	return requisition, nil
}

func validateRequestedItems(items []*RequisitionItem, requestedItems []*RequestedItem) error {
	if requestedItems == nil {
		return nil
	}
	available := make(map[string]*RequestedItem, len(requestedItems))
	for _, requestedItem := range requestedItems {
		if requestedItem == nil {
			return ErrNilFields("requestedItem")
		}
		if _, exists := available[requestedItem.ID()]; exists {
			return ErrDuplicateRequestedItem
		}
		available[requestedItem.ID()] = requestedItem
	}
	for _, item := range items {
		if item.RequestedItemID() == "" {
			continue
		}
		if _, exists := available[item.RequestedItemID()]; !exists {
			return ErrInvalidRequestedItemID
		}
	}
	return nil
}

func newRequisitionHeader(
	id string,
	requestedBy string,
	departmentID string,
	priority RequisitionPriority,
	neededBy time.Time,
	justification *RequisitionJustification,
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
	if !priority.IsValid() {
		return nil, ErrInvalidRequisitionPriority
	}
	if justification == nil {
		return nil, ErrNilFields("justification")
	}
	if isDateNotFuture(neededBy) {
		return nil, ErrInvalidNeededByDate
	}
	now := time.Now()
	return &Requisition{
		id: id, requestedBy: requestedBy, departmentID: departmentID,
		priority: priority, neededBy: neededBy, justification: justification,
		status: RequisitionStatusPending, createdAt: now, updatedAt: now,
	}, nil
}

func validateRequisitionItems(requisitionID string, items []*RequisitionItem) error {
	seenItems := make(map[string]struct{}, len(items))
	seenRequestedItems := make(map[string]struct{}, len(items))
	seenLines := make(map[int]struct{}, len(items))
	for _, item := range items {
		if item == nil {
			return ErrNilFields("requisitionItem")
		}
		if item.RequisitionID() != requisitionID {
			return ErrInvalidRequisitionItemID
		}
		if _, exists := seenLines[item.LineNumber()]; exists {
			return ErrInvalidLineNumber
		}
		seenLines[item.LineNumber()] = struct{}{}
		if item.ItemID() != "" {
			if _, exists := seenItems[item.ItemID()]; exists {
				return ErrDuplicateRequisitionItem
			}
			seenItems[item.ItemID()] = struct{}{}
		}
		if item.RequestedItemID() != "" {
			if _, exists := seenRequestedItems[item.RequestedItemID()]; exists {
				return ErrDuplicateRequestedItem
			}
			seenRequestedItems[item.RequestedItemID()] = struct{}{}
		}
	}
	return nil
}

func isDateNotFuture(value time.Time) bool {
	today := time.Now().UTC()
	date := value.UTC()
	if date.Year() != today.Year() {
		return date.Year() <= today.Year()
	}
	if date.Month() != today.Month() {
		return date.Month() <= today.Month()
	}
	return date.Day() <= today.Day()
}

func (r *Requisition) ReferenceCode() string { return r.referenceCode }

func (r *Requisition) Items() []*RequisitionItem {
	return append([]*RequisitionItem(nil), r.items...)
}

func (r *Requisition) RequestedItems() []*RequestedItem {
	return append([]*RequestedItem(nil), r.requestedItems...)
}

func (r *Requisition) StatusHistory() []*RequisitionStatusHistory {
	return append([]*RequisitionStatusHistory(nil), r.statusHistory...)
}

func (r *Requisition) IsEditable() bool {
	return r.status == RequisitionStatusPending
}

func (r *Requisition) TotalEstimatedAmount() (*shareddomain.Money, error) {
	if len(r.items) == 0 {
		return shareddomain.NewNairaMoney(0)
	}
	var total *shareddomain.Money
	for _, item := range r.items {
		lineTotal, err := item.LineTotal()
		if err != nil {
			return nil, err
		}
		if total == nil {
			total = lineTotal
			continue
		}
		total, err = total.Add(lineTotal)
		if err != nil {
			return nil, err
		}
	}
	return total, nil
}

func (r *Requisition) InitialStatusHistory(
	id string,
	actedBy string,
	actorRole RequisitionActorRole,
	createdAt time.Time,
) (*RequisitionStatusHistory, error) {
	return NewRequisitionStatusHistory(
		id, r.id, 1, nil, RequisitionStatusPending,
		actedBy, actorRole, "", createdAt,
	)
}

func (r *Requisition) RecordInitialStatusHistory(history *RequisitionStatusHistory) error {
	if history == nil {
		return ErrNilFields("statusHistory")
	}
	if history.RequisitionID() != r.id || history.Sequence() != 1 ||
		history.FromStatus() != nil || history.ToStatus() != RequisitionStatusPending {
		return ErrInvalidStatusSequence
	}
	r.statusHistory = append(r.statusHistory, history)
	return nil
}

func (r *Requisition) ApplyStatusTransition(
	id string,
	sequence int,
	newStatus RequisitionStatus,
	actedBy string,
	actorRole RequisitionActorRole,
	reason string,
	createdAt time.Time,
) (*RequisitionStatusHistory, error) {
	if !r.status.CanTransitionTo(newStatus) {
		return nil, ErrCannotTransitionToStatus(r.status.String(), newStatus.String())
	}
	fromStatus := r.status
	history, err := NewRequisitionStatusHistory(
		id, r.id, sequence, &fromStatus, newStatus,
		actedBy, actorRole, reason, createdAt,
	)
	if err != nil {
		return nil, err
	}
	r.status = newStatus
	r.updatedAt = createdAt
	r.statusHistory = append(r.statusHistory, history)
	return history, nil
}

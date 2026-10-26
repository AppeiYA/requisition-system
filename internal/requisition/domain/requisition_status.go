package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

type RequisitionStatus string

const (
	RequisitionStatusPending       RequisitionStatus = "pending"
	RequisitionStatusApproved      RequisitionStatus = "approved"
	RequisitionStatusRejected      RequisitionStatus = "rejected"
	RequisitionStatusCompleted     RequisitionStatus = "completed"
	RequisitionStatusCancelled     RequisitionStatus = "cancelled"
	RequisitionStatusInFulfillment RequisitionStatus = "in_fulfillment"
)

var (
	ErrInvalidRequisitionStatus = customerrors.NewBadRequestError("invalid requisition status")
)

var validRequisitionStatusTransitions = map[RequisitionStatus][]RequisitionStatus{
	RequisitionStatusPending:       {RequisitionStatusApproved, RequisitionStatusRejected, RequisitionStatusCancelled},
	RequisitionStatusApproved:      {RequisitionStatusInFulfillment, RequisitionStatusCancelled},
	RequisitionStatusRejected:      {},
	RequisitionStatusCompleted:     {},
	RequisitionStatusCancelled:     {},
	RequisitionStatusInFulfillment: {RequisitionStatusCompleted, RequisitionStatusCancelled},
}

func (rs RequisitionStatus) CanTransitionTo(newStatus RequisitionStatus) bool {
	validTransitions, exists := validRequisitionStatusTransitions[rs]
	if !exists {
		return false
	}

	for _, validStatus := range validTransitions {
		if newStatus == validStatus {
			return true
		}
	}

	return false
}
func (rs RequisitionStatus) IsValid() bool {
	switch rs {
	case RequisitionStatusPending, RequisitionStatusApproved, RequisitionStatusRejected, RequisitionStatusCompleted, RequisitionStatusCancelled, RequisitionStatusInFulfillment:
		return true
	default:
		return false
	}
}

func NewRequisitionStatus(value string) (RequisitionStatus, error) {
	status := RequisitionStatus(value)
	if !status.IsValid() {
		return "", ErrInvalidRequisitionStatus
	}
	return status, nil
}

func (rs RequisitionStatus) String() string {
	return string(rs)
}

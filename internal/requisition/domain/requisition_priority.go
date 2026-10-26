package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

type RequisitionPriority string

const (
	PriorityLow    RequisitionPriority = "low"
	PriorityNormal RequisitionPriority = "normal"
	PriorityMedium RequisitionPriority = "medium"
	PriorityUrgent RequisitionPriority = "urgent"
)

var (
	ErrInvalidRequisitionPriority = customerrors.NewBadRequestError("invalid requisition priority")
)

func (rp RequisitionPriority) IsValid() bool {
	switch rp {
	case PriorityLow, PriorityNormal, PriorityMedium, PriorityUrgent:
		return true
	default:
		return false
	}
}

func NewRequisitionPriority(value string) (RequisitionPriority, error) {
	priority := RequisitionPriority(value)
	if !priority.IsValid() {
		return "", ErrInvalidRequisitionPriority
	}
	return priority, nil
}

func (rp RequisitionPriority) String() string {
	return string(rp)
}

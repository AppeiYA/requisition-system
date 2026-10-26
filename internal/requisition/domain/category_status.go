package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

type CategoryStatus string

const (
	CategoryStatusActive   CategoryStatus = "active"
	CategoryStatusInactive CategoryStatus = "inactive"
)

var (
	ErrInvalidCategoryStatus = customerrors.NewBadRequestError("invalid category status")
)

var validTransitions = map[CategoryStatus][]CategoryStatus{
	CategoryStatusActive:   {CategoryStatusInactive},
	CategoryStatusInactive: {CategoryStatusActive},
}

func (c CategoryStatus) CanTransitionTo(newStatus CategoryStatus) bool {
	validNextStatuses, exists := validTransitions[c]
	if !exists {
		return false
	}
	for _, status := range validNextStatuses {
		if status == newStatus {
			return true
		}
	}
	return false
}

func (c CategoryStatus) IsValid() bool {
	switch c {
	case CategoryStatusActive, CategoryStatusInactive:
		return true
	default:
		return false
	}
}

func NewCategoryStatus(value string) (*CategoryStatus, error) {
	status := CategoryStatus(value)
	if !status.IsValid() {
		return nil, ErrInvalidCategoryStatus
	}
	return &status, nil
}

func (c CategoryStatus) String() string {
	return string(c)
}

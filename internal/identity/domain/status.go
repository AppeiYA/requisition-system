package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/errors"

var (
	ErrInvalidStatus = customerrors.NewBadRequestError("Invalid status")
)

type Status string

const (
	ACTIVE    Status = "active"
	SUSPENDED Status = "suspended"
)

var validTransitions = map[Status][]Status{
	ACTIVE:    {SUSPENDED},
	SUSPENDED: {ACTIVE},
}

func (s Status) CanTransitionTo(newStatus Status) bool {
	validNextStatuses, ok := validTransitions[s]
	if !ok {
		return false
	}
	for _, validStatus := range validNextStatuses {
		if newStatus == validStatus {
			return true
		}
	}
	return false
}

func (s Status) IsValid() bool {
	switch s {
	case ACTIVE, SUSPENDED:
		return true
	default:
		return false
	}
}

func NewStatus(value string) (Status, error) {
	s := Status(value)
	if !s.IsValid() {
		return "", ErrInvalidStatus
	}
	return s, nil
}

func (s Status) Value() string {
	return string(s)
}

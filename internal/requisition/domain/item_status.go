package domain

import (
	"strings"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

type ItemStatus string

const (
	ItemStatusActive   ItemStatus = "active"
	ItemStatusInactive ItemStatus = "inactive"
)

var (
	ErrInvalidItemStatus = customerrors.NewBadRequestError("invalid item status")
)

var validItemStatusTransitions = map[ItemStatus][]ItemStatus{
	ItemStatusActive:   {ItemStatusInactive},
	ItemStatusInactive: {ItemStatusActive},
}

func (is ItemStatus) CanTransitionTo(newStatus ItemStatus) bool {
	validTransitions, ok := validItemStatusTransitions[is]
	if !ok {
		return false
	}

	for _, validTransition := range validTransitions {
		if validTransition == newStatus {
			return true
		}
	}

	return false

}

func (is ItemStatus) IsValid() bool {
	switch is {
	case ItemStatusActive, ItemStatusInactive:
		return true
	default:
		return false
	}
}

func NewItemStatus(status string) (ItemStatus, error) {
	itemStatus := ItemStatus(strings.Trim(status, ""))
	if !itemStatus.IsValid() {
		return "", ErrInvalidItemStatus
	}

	return itemStatus, nil
}

func (is ItemStatus) String() string {
	return string(is)
}

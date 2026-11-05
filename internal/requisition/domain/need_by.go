package domain

import (
	"time"
)

type NeedBy struct {
	value *time.Time
}

var ErrInvalidNeedByDate = ErrInvalidNeededByDate

func NewNeedBy(value *time.Time) (*NeedBy, error) {
	if isDateNotFuture(value) {
		return nil, ErrInvalidNeedByDate
	}
	return &NeedBy{value: value}, nil
}

func (n *NeedBy) Value() *time.Time {
	return n.value
}

func isDateNotFuture(value *time.Time) bool {
	if value == nil {
		return false
	}
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

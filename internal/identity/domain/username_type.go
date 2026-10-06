package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/errors"

type UsernameType string

const (
	FirstName UsernameType = "first_name"
	LastName  UsernameType = "last_name"
)

func (u UsernameType) isValid() bool {
	switch u {
	case FirstName, LastName:
		return true
	default:
		return false
	}
}

var (
	ErrInvalidUsernameType = customerrors.NewInternalServerError("invalid username type")
)

func NewUsernameType(value string) (UsernameType, error) {
	nameType := UsernameType(value)
	if !nameType.isValid() {
		return "", ErrInvalidUsernameType
	}
	return nameType, nil
}

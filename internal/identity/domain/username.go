package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/errors"

var (
	ErrUsernameTooShort = func(nameType UsernameType) error {
		return customerrors.NewBadRequestError(string(nameType) + " is too short")
	}
	ErrUsernameTooLong = func(nameType UsernameType) error {
		return customerrors.NewBadRequestError(string(nameType) + " is too long")
	}
	ErrInvalidUsername = func(nameType UsernameType) error {
		return customerrors.NewBadRequestError("Invalid " + string(nameType))
	}
)

type Username struct {
	value string
}

func NewUsername(value string, nameType UsernameType) (*Username, error) {
	if len(value) < 2 {
		return nil, ErrUsernameTooShort(nameType)
	}
	if len(value) > 100 {
		return nil, ErrUsernameTooLong(nameType)
	}
	if !nameType.isValid() {
		return nil, ErrInvalidUsername(nameType)
	}
	return &Username{value: value}, nil
}

func (u *Username) Value() string {
	return u.value
}

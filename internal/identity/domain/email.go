package domain

import (
	"net/mail"
	"regexp"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/errors"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
var (
	ErrInvalidEmailFormat = customerrors.NewBadRequestError("Invalid email address")
)

type Email struct {
	value string
}

func NewEmail(value string) (*Email, error) {
	if !isValidEmail(value) {
		return nil, ErrInvalidEmailFormat
	}
	return &Email{value: value}, nil
}

func (e *Email) Value() string {
	return e.value
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	if len(email) > 254 {
		return false
	}
	if ok := emailRegex.MatchString(email); !ok {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

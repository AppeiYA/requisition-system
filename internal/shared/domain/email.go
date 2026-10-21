package shareddomain

import (
	"net/mail"
	"strings"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

type Email struct {
	value string
}

var (
	ErrInvalidEmailFormat = customerrors.NewBadRequestError("Invalid email address")
)

func NewEmail(value string) (*Email, error) {
	value = strings.ToLower(value)
	if !validateEmail(value) {
		return nil, ErrInvalidEmailFormat
	}
	return &Email{value: value}, nil
}

func (e *Email) Value() string {
	return e.value
}

func (e *Email) IsZero() bool {
	return e.value == ""
}

func (e *Email) Equals(other *Email) bool {
	if other == nil || e == nil {
		return false
	}
	return e.value == other.value
}

func validateEmail(value string) bool {
	if value == "" || len(value) > 254 {
		return false
	}

	if strings.TrimSpace(value) != value {
		return false
	}

	parsed, err := mail.ParseAddress(value)
	if err != nil {
		return false
	}

	if parsed.Address != value {
		return false
	}

	// Require a dotted domain with a top-level label, e.g. example.com.
	at := strings.LastIndex(value, "@")
	domain := value[at+1:]
	if !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	return true
}

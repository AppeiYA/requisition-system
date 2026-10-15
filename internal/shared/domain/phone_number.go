package shareddomain

import (
	"regexp"
	"strings"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrInvalidPhoneNumber = customerrors.NewBadRequestError("invalid phone number")
)

// phoneNumberPattern matches E.164 formatted numbers, e.g. "+14155551234".
var phoneNumberPattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

type PhoneNumber struct {
	value string
}

// NewPhoneNumber validates and normalizes a phone number. The value must be in
// E.164 format (with a leading "+").
func NewPhoneNumber(value string) (*PhoneNumber, error) {
	normalized := normalizePhoneNumber(value)
	if !phoneNumberPattern.MatchString(normalized) {
		return nil, ErrInvalidPhoneNumber
	}
	return &PhoneNumber{value: normalized}, nil
}

// NewOptionalPhoneNumber behaves like NewPhoneNumber but treats an empty value
// as "not provided", returning a nil PhoneNumber and no error.
func NewOptionalPhoneNumber(value string) (*PhoneNumber, error) {
	if normalizePhoneNumber(value) == "" {
		return nil, nil
	}
	return NewPhoneNumber(value)
}

// normalizePhoneNumber removes common separators so that formatting differences
// do not affect validation.
func normalizePhoneNumber(value string) string {
	replacer := strings.NewReplacer(
		" ", "",
		"-", "",
		"(", "",
		")", "",
		".", "",
	)
	return replacer.Replace(strings.TrimSpace(value))
}

func (p *PhoneNumber) Value() string {
	if p == nil {
		return ""
	}
	return p.value
}

func (p *PhoneNumber) IsZero() bool {
	return p == nil || p.value == ""
}

func (p *PhoneNumber) Equals(other *PhoneNumber) bool {
	if p == nil || other == nil {
		return p == other
	}
	return p.value == other.value
}

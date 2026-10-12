package domain

import (
	"strings"
	"unicode"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/errors"
)

var (
	ErrPasswordIsRequired = customerrors.NewBadRequestError("password is required")
	ErrInvalidPassword    = customerrors.NewBadRequestError("password must be at least 8 characters long, contain at least one uppercase letter, one lowercase letter, one number and one special character")
)

type Password struct {
	value string
}

func NewPassword(value string) (*Password, error) {
	if value == "" {
		return nil, ErrPasswordIsRequired
	}
	
	if !isValidPassword(value) {
		return nil, ErrInvalidPassword
	}
	
	return &Password{value: value}, nil
}

func (p *Password) Value() string {
	return p.value
}

// isValidPassword handles the complexity checks efficiently without crashing Go
func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool
	specialChars := "@$!%*?&"

	for _, ch := range password {
		switch {
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case strings.ContainsRune(specialChars, ch):
			hasSpecial = true
		}
	}

	return hasLower && hasUpper && hasDigit && hasSpecial
}

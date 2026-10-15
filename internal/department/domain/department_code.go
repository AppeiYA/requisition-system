package domain

import (
	"strings"
	"unicode"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

type DepartmentCode struct {
	value string
}

var (
	ErrEmptyDepartmentCode   = customerrors.NewBadRequestError("Department code cannot be empty")
	ErrDepartmentCodeTooLong = customerrors.NewBadRequestError("Department code cannot exceed 4 characters")
	ErrDepartmentCodeInvalid = customerrors.NewBadRequestError("Department code must contain only letters")
)

func NewDepartmentCode(value string) (*DepartmentCode, error) {
	if value == "" {
		return nil, ErrEmptyDepartmentCode
	}
	if len(value) > 4 {
		return nil, ErrDepartmentCodeTooLong
	}

	for _, r := range value {
		if !unicode.IsLetter(r) {
			return nil, ErrDepartmentCodeInvalid
		}
	}

	return &DepartmentCode{value: strings.ToUpper(value)}, nil
}

func (dc *DepartmentCode) Value() string {
	return dc.value
}

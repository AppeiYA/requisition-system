package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

type DepartmentName struct {
	value string
}

var (
	ErrEmptyDepartmentName = customerrors.NewBadRequestError("Department name cannot be empty")
	ErrDepartmentNameTooLong = customerrors.NewBadRequestError("Department name cannot exceed 60 characters")
)

func NewDepartmentName(value string) (*DepartmentName, error) {
	if value == "" {
		return nil, ErrEmptyDepartmentName
	}
	if len(value) > 60 {
		return nil, ErrDepartmentNameTooLong
	}

	return &DepartmentName{value: value}, nil
}

func (dn *DepartmentName) Value() string {
	return dn.value
}
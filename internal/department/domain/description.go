package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

type DepartmentDescription struct {
	value string
}

var (
	ErrEmptyDepartmentDescription = customerrors.NewBadRequestError("Department description cannot be empty")
	ErrDepartmentDescriptionTooLong = customerrors.NewBadRequestError("Department description cannot exceed 255 characters")
)

func NewDepartmentDescription(value string) (*DepartmentDescription, error) {
	if value == "" {
		return nil, ErrEmptyDepartmentDescription
	}
	if len(value) > 255 {
		return nil, ErrDepartmentDescriptionTooLong
	}

	return &DepartmentDescription{value: value}, nil
}

func (dd *DepartmentDescription) Value() string {
	return dd.value
}
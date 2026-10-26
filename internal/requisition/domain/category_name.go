package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

var (
	ErrCategoryNameEmpty   = customerrors.NewBadRequestError("category name cannot be empty")
	ErrCategoryNameTooLong = customerrors.NewBadRequestError("category name cannot be longer than 100 characters")
)

type CategoryName struct {
	value string
}

func NewCategoryName(value string) (*CategoryName, error) {
	if value == "" {
		return nil, ErrCategoryNameEmpty
	}
	if len(value) > 100 {
		return nil, ErrCategoryNameTooLong
	}

	return &CategoryName{
		value: value,
	}, nil
}

func (c *CategoryName) Value() string {
	return c.value
}

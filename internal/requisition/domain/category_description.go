package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

type CategoryDescription struct {
	value string
}

var (
	ErrCategoryDescriptionTooLong = customerrors.NewBadRequestError("category description cannot be longer than 200 characters")
)

func NewCategoryDescription(value string) (*CategoryDescription, error) {
	if len(value) > 200 {
		return nil, ErrCategoryDescriptionTooLong
	}
	return &CategoryDescription{
		value: value,
	}, nil
}

func (c *CategoryDescription) Value() string {
	return c.value
}

package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

type ItemDescription struct {
	value string
}

var (
	ErrItemDescriptionCannotBeEmpty = customerrors.NewBadRequestError("item description cannot be empty")
	ErrItemDescriptionTooLong       = customerrors.NewBadRequestError("item description is too long")
)

func NewItemDescription(description string) (*ItemDescription, error) {
	if description == "" {
		return nil, ErrItemDescriptionCannotBeEmpty
	}

	if len(description) > 200 {
		return nil, ErrItemDescriptionTooLong
	}

	return &ItemDescription{value: description}, nil
}

func (i *ItemDescription) Value() string {
	return i.value
}

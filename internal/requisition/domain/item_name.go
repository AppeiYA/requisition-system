package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

type ItemName struct {
	value string
}

var (
	ErrItemNameCannotBeEmpty = customerrors.NewBadRequestError("item name cannot be empty")
	ErrItemNameTooLong       = customerrors.NewBadRequestError("item name cannot be longer than 100 characters")
)

func NewItemName(name string) (*ItemName, error) {
	if name == "" {
		return nil, ErrItemNameCannotBeEmpty
	}

	if len(name) > 100 {
		return nil, ErrItemNameTooLong
	}

	return &ItemName{value: name}, nil
}

func (n *ItemName) Value() string {
	return n.value
}

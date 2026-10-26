package domain

import (
	"time"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var CategoryPrefixID = "cat"

type ItemCategory struct {
	id          string
	name        *CategoryName
	description *CategoryDescription
	status      CategoryStatus
	createdAt   time.Time
}

var (
	ErrInvalidCategoryID      = customerrors.NewBadRequestError("invalid category ID")
	ErrNoCategoryUpdateFields = customerrors.NewBadRequestError("at least one of name or description must be provided")
)

func NewItemCategory(id string, name *CategoryName, description *CategoryDescription) (*ItemCategory, error) {
	if id == "" {
		return nil, ErrInvalidCategoryID
	}

	status := CategoryStatusActive

	now := time.Now()

	return &ItemCategory{
		id:          id,
		name:        name,
		description: description,
		status:      status,
		createdAt:   now,
	}, nil
}

func ReconstituteItemCategory(id string, name string, description string, status string, createdAt time.Time) (*ItemCategory, error) {
	categoryName, err := NewCategoryName(name)
	if err != nil {
		return nil, err
	}

	categoryDescription, err := NewCategoryDescription(description)
	if err != nil {
		return nil, err
	}

	categoryStatus, err := NewCategoryStatus(status)
	if err != nil {
		return nil, err
	}

	return &ItemCategory{
		id:          id,
		name:        categoryName,
		description: categoryDescription,
		status:      *categoryStatus,
		createdAt:   createdAt,
	}, nil
}

func (c *ItemCategory) ID() string {
	return c.id
}

func (c *ItemCategory) Name() *CategoryName {
	return c.name
}

func (c *ItemCategory) Description() *CategoryDescription {
	return c.description
}

func (c *ItemCategory) Status() CategoryStatus {
	return c.status
}

func (c *ItemCategory) CreatedAt() time.Time {
	return c.createdAt
}

// Update applies a partial update to the category name and description. A nil
// field is left unchanged. At least one field must be provided. Status changes
// go through Activate and Deactivate instead.
func (c *ItemCategory) Update(name *string, description *string) error {
	if name == nil && description == nil {
		return ErrNoCategoryUpdateFields
	}

	var (
		newName        *CategoryName
		newDescription *CategoryDescription
	)

	if name != nil {
		value, err := NewCategoryName(*name)
		if err != nil {
			return err
		}
		newName = value
	}

	if description != nil {
		value, err := NewCategoryDescription(*description)
		if err != nil {
			return err
		}
		newDescription = value
	}

	if newName != nil {
		c.name = newName
	}
	if newDescription != nil {
		c.description = newDescription
	}

	return nil
}

func (c *ItemCategory) Activate() {
	if c.status.CanTransitionTo(CategoryStatusActive) {
		c.status = CategoryStatusActive
	}
}

func (c *ItemCategory) Deactivate() {
	if c.status.CanTransitionTo(CategoryStatusInactive) {
		c.status = CategoryStatusInactive
	}
}

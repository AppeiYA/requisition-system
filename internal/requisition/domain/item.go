package domain

import (
	"strings"
	"time"

	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var ItemPrefixID = "itm"

type Item struct {
	id          string
	name        *ItemName
	description *ItemDescription
	unitCost    *shareddomain.Money
	categoryID  string
	status      ItemStatus
	createdAt   time.Time
}

var (
	ErrInvalidItemID      = customerrors.NewBadRequestError("invalid item ID")
	ErrNoItemUpdateFields = customerrors.NewBadRequestError("at least one field must be provided to update an item")
)

func NewItem(id string, name *ItemName, description *ItemDescription, categoryID string, unitCost float64) (*Item, error) {
	if id == "" || !strings.HasPrefix(id, ItemPrefixID) {
		return nil, ErrInvalidItemID
	}
	if categoryID == "" || !strings.HasPrefix(categoryID, CategoryPrefixID) {
		return nil, ErrInvalidCategoryID
	}

	status := ItemStatusActive

	newUnitCost, err := shareddomain.NewNairaMoney(unitCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &Item{
		id:          id,
		name:        name,
		description: description,
		categoryID:  categoryID,
		unitCost:    newUnitCost,
		status:      status,
		createdAt:   now,
	}, nil
}

func ReconstituteItem(id string, name string, description string, categoryID string, unitCostKobo int64, status string, createdAt time.Time) (*Item, error) {
	if id == "" || !strings.HasPrefix(id, ItemPrefixID) {
		return nil, ErrInvalidItemID
	}

	if categoryID == "" || !strings.HasPrefix(categoryID, CategoryPrefixID) {
		return nil, ErrInvalidCategoryID
	}

	itemStatus, err := NewItemStatus(status)
	if err != nil {
		return nil, err
	}

	itemName, err := NewItemName(name)
	if err != nil {
		return nil, err
	}

	itemDescription, err := NewItemDescription(description)
	if err != nil {
		return nil, err
	}

	unitCost, err := shareddomain.NewNairaMoneyFromMinorUnits(unitCostKobo)
	if err != nil {
		return nil, err
	}

	return &Item{
		id:          id,
		name:        itemName,
		description: itemDescription,
		categoryID:  categoryID,
		unitCost:    unitCost,
		status:      itemStatus,
		createdAt:   createdAt,
	}, nil
}

func (i *Item) ID() string {
	return i.id
}

func (i *Item) Name() *ItemName {
	return i.name
}

func (i *Item) Description() *ItemDescription {
	return i.description
}

func (i *Item) CategoryID() string {
	return i.categoryID
}

func (i *Item) Status() ItemStatus {
	return i.status
}

func (i *Item) UnitCost() *shareddomain.Money {
	return i.unitCost
}

func (i *Item) CreatedAt() time.Time {
	return i.createdAt
}

// Update applies a partial update to the item. A nil field is left unchanged.
// At least one field must be provided. Status changes go through Activate and
// Deactivate instead.
func (i *Item) Update(
	name *string,
	description *string,
	categoryID *string,
	unitCost *float64,
) error {
	if name == nil && description == nil && categoryID == nil && unitCost == nil {
		return ErrNoItemUpdateFields
	}

	var (
		newName        *ItemName
		newDescription *ItemDescription
		newCategoryID  string
		newUnitCost    *shareddomain.Money
	)

	if name != nil {
		value, err := NewItemName(*name)
		if err != nil {
			return err
		}
		newName = value
	}

	if description != nil {
		value, err := NewItemDescription(*description)
		if err != nil {
			return err
		}
		newDescription = value
	}

	if categoryID != nil {
		if *categoryID == "" || !strings.HasPrefix(*categoryID, CategoryPrefixID) {
			return ErrInvalidCategoryID
		}
		newCategoryID = *categoryID
	}

	if unitCost != nil {
		value, err := shareddomain.NewNairaMoney(*unitCost)
		if err != nil {
			return err
		}
		newUnitCost = value
	}

	if newName != nil {
		i.name = newName
	}
	if newDescription != nil {
		i.description = newDescription
	}
	if categoryID != nil {
		i.categoryID = newCategoryID
	}
	if newUnitCost != nil {
		i.unitCost = newUnitCost
	}

	return nil
}

func (i *Item) Activate() error {
	if !i.status.CanTransitionTo(ItemStatusActive) {
		return customerrors.NewBadRequestError("cannot activate item from current status")
	}

	i.status = ItemStatusActive
	return nil
}

func (i *Item) Deactivate() error {
	if !i.status.CanTransitionTo(ItemStatusInactive) {
		return customerrors.NewBadRequestError("cannot deactivate item from current status")
	}

	i.status = ItemStatusInactive
	return nil
}

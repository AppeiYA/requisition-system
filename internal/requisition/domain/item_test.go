package domain

import (
	"math"
	"strings"
	"testing"
	"time"

	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

func Test_NewItem(t *testing.T) {
	t.Parallel()

	name := mustItemName(t, "Laptop")
	description := mustItemDescription(t, "Business laptop")

	before := time.Now()
	item, err := NewItem("itm_123", name, description, "cat_123", 123.45)
	after := time.Now()

	if err != nil {
		t.Fatalf("NewItem() unexpected error = %v", err)
	}
	if item == nil {
		t.Fatal("NewItem() = nil, want item")
	}
	if item.ID() != "itm_123" {
		t.Errorf("Item.ID() = %q, want %q", item.ID(), "itm_123")
	}
	if item.Name() != name {
		t.Errorf("Item.Name() = %v, want %v", item.Name(), name)
	}
	if item.Description() != description {
		t.Errorf("Item.Description() = %v, want %v", item.Description(), description)
	}
	if item.CategoryID() != "cat_123" {
		t.Errorf("Item.CategoryID() = %q, want %q", item.CategoryID(), "cat_123")
	}
	if item.Status() != ItemStatusActive {
		t.Errorf("Item.Status() = %q, want %q", item.Status(), ItemStatusActive)
	}
	if item.CreatedAt().Before(before) || item.CreatedAt().After(after) {
		t.Errorf("Item.CreatedAt() = %v, expected between %v and %v", item.CreatedAt(), before, after)
	}
}

func Test_NewItem_Validation(t *testing.T) {
	t.Parallel()

	name := mustItemName(t, "Laptop")
	description := mustItemDescription(t, "Business laptop")

	tests := []struct {
		name       string
		id         string
		categoryID string
		unitCost   float64
		wantErr    error
	}{
		{
			name:       "empty id",
			id:         "",
			categoryID: "cat_123",
			unitCost:   100,
			wantErr:    ErrInvalidItemID,
		},
		{
			name:       "invalid id prefix",
			id:         "item_123",
			categoryID: "cat_123",
			unitCost:   100,
			wantErr:    ErrInvalidItemID,
		},
		{
			name:       "empty category id",
			id:         "itm_123",
			categoryID: "",
			unitCost:   100,
			wantErr:    ErrInvalidCategoryID,
		},
		{
			name:       "invalid category id prefix",
			id:         "itm_123",
			categoryID: "dept_123",
			unitCost:   100,
			wantErr:    ErrInvalidCategoryID,
		},
		{
			name:       "negative unit cost",
			id:         "itm_123",
			categoryID: "cat_123",
			unitCost:   -1,
			wantErr:    shareddomain.ErrNegativeAmount,
		},
		{
			name:       "NaN unit cost",
			id:         "itm_123",
			categoryID: "cat_123",
			unitCost:   math.NaN(),
			wantErr:    shareddomain.ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			item, err := NewItem(tt.id, name, description, tt.categoryID, tt.unitCost)

			if err != tt.wantErr {
				t.Fatalf("NewItem() error = %v, want %v", err, tt.wantErr)
			}
			if item != nil {
				t.Errorf("NewItem() = %v, want nil", item)
			}
		})
	}
}

func Test_ReconstituteItem(t *testing.T) {
	t.Parallel()

	createdAt := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name        string
		id          string
		itemName    string
		description string
		categoryID  string
		status      string
		wantStatus  ItemStatus
		wantErr     error
	}{
		{
			name:        "valid active item",
			id:          "itm_123",
			itemName:    "Laptop",
			description: "Business laptop",
			categoryID:  "cat_123",
			status:      "active",
			wantStatus:  ItemStatusActive,
		},
		{
			name:        "valid inactive item",
			id:          "itm_123",
			itemName:    "Laptop",
			description: "Business laptop",
			categoryID:  "cat_123",
			status:      "inactive",
			wantStatus:  ItemStatusInactive,
		},
		{
			name:        "empty id",
			id:          "",
			itemName:    "Laptop",
			description: "Business laptop",
			categoryID:  "cat_123",
			status:      "active",
			wantErr:     ErrInvalidItemID,
		},
		{
			name:        "invalid id prefix",
			id:          "item_123",
			itemName:    "Laptop",
			description: "Business laptop",
			categoryID:  "cat_123",
			status:      "active",
			wantErr:     ErrInvalidItemID,
		},
		{
			name:        "empty category id",
			id:          "itm_123",
			itemName:    "Laptop",
			description: "Business laptop",
			categoryID:  "",
			status:      "active",
			wantErr:     ErrInvalidCategoryID,
		},
		{
			name:        "invalid category id prefix",
			id:          "itm_123",
			itemName:    "Laptop",
			description: "Business laptop",
			categoryID:  "dept_123",
			status:      "active",
			wantErr:     ErrInvalidCategoryID,
		},
		{
			name:        "invalid status",
			id:          "itm_123",
			itemName:    "Laptop",
			description: "Business laptop",
			categoryID:  "cat_123",
			status:      "archived",
			wantErr:     ErrInvalidItemStatus,
		},
		{
			name:        "invalid item name",
			id:          "itm_123",
			itemName:    "",
			description: "Business laptop",
			categoryID:  "cat_123",
			status:      "active",
			wantErr:     ErrItemNameCannotBeEmpty,
		},
		{
			name:        "item name too long",
			id:          "itm_123",
			itemName:    strings.Repeat("a", 101),
			description: "Business laptop",
			categoryID:  "cat_123",
			status:      "active",
			wantErr:     ErrItemNameTooLong,
		},
		{
			name:        "invalid item description",
			id:          "itm_123",
			itemName:    "Laptop",
			description: "",
			categoryID:  "cat_123",
			status:      "active",
			wantErr:     ErrItemDescriptionCannotBeEmpty,
		},
		{
			name:        "item description too long",
			id:          "itm_123",
			itemName:    "Laptop",
			description: strings.Repeat("a", 201),
			categoryID:  "cat_123",
			status:      "active",
			wantErr:     ErrItemDescriptionTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			item, err := ReconstituteItem(
				tt.id,
				tt.itemName,
				tt.description,
				tt.categoryID,
				tt.status,
				createdAt,
			)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("ReconstituteItem() error = %v, want %v", err, tt.wantErr)
				}
				if item != nil {
					t.Errorf("ReconstituteItem() = %v, want nil", item)
				}
				return
			}

			if err != nil {
				t.Fatalf("ReconstituteItem() unexpected error = %v", err)
			}
			if item == nil {
				t.Fatal("ReconstituteItem() = nil, want item")
			}
			if item.ID() != tt.id {
				t.Errorf("Item.ID() = %q, want %q", item.ID(), tt.id)
			}
			if item.Name().Value() != tt.itemName {
				t.Errorf("Item.Name().Value() = %q, want %q", item.Name().Value(), tt.itemName)
			}
			if item.Description().Value() != tt.description {
				t.Errorf("Item.Description().Value() = %q, want %q", item.Description().Value(), tt.description)
			}
			if item.CategoryID() != tt.categoryID {
				t.Errorf("Item.CategoryID() = %q, want %q", item.CategoryID(), tt.categoryID)
			}
			if item.Status() != tt.wantStatus {
				t.Errorf("Item.Status() = %q, want %q", item.Status(), tt.wantStatus)
			}
			if !item.CreatedAt().Equal(createdAt) {
				t.Errorf("Item.CreatedAt() = %v, want %v", item.CreatedAt(), createdAt)
			}
		})
	}
}

func Test_Item_ActivateAndDeactivate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		initial    ItemStatus
		activate   bool
		wantStatus ItemStatus
		wantErr    bool
	}{
		{
			name:       "active item can be deactivated",
			initial:    ItemStatusActive,
			activate:   false,
			wantStatus: ItemStatusInactive,
			wantErr:    false,
		},
		{
			name:       "inactive item can be activated",
			initial:    ItemStatusInactive,
			activate:   true,
			wantStatus: ItemStatusActive,
			wantErr:    false,
		},
		{
			name:       "active item cannot be activated again",
			initial:    ItemStatusActive,
			activate:   true,
			wantStatus: ItemStatusActive,
			wantErr:    true,
		},
		{
			name:       "inactive item cannot be deactivated again",
			initial:    ItemStatusInactive,
			activate:   false,
			wantStatus: ItemStatusInactive,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			item := mustItemWithStatus(t, tt.initial)

			var err error
			if tt.activate {
				err = item.Activate()
			} else {
				err = item.Deactivate()
			}

			if tt.wantErr {
				if err == nil {
					t.Fatal("Item status update error = nil, want error")
				}
			} else if err != nil {
				t.Fatalf("Item status update unexpected error = %v", err)
			}

			if item.Status() != tt.wantStatus {
				t.Errorf("Item.Status() = %q, want %q", item.Status(), tt.wantStatus)
			}
		})
	}
}

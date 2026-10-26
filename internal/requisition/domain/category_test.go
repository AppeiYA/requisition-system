package domain

import (
	"strings"
	"testing"
	"time"
)

func Test_NewItemCategory(t *testing.T) {
	t.Parallel()

	name := mustCategoryName(t, "Office Supplies")
	description := mustCategoryDescription(t, "General office supplies")

	before := time.Now()
	category, err := NewItemCategory("cat_123", name, description)
	after := time.Now()

	if err != nil {
		t.Fatalf("NewItemCategory() unexpected error = %v", err)
	}
	if category == nil {
		t.Fatal("NewItemCategory() = nil, want category")
	}
	if category.ID() != "cat_123" {
		t.Errorf("ItemCategory.ID() = %q, want %q", category.ID(), "cat_123")
	}
	if category.Name() != name {
		t.Errorf("ItemCategory.Name() = %v, want %v", category.Name(), name)
	}
	if category.Description() != description {
		t.Errorf("ItemCategory.Description() = %v, want %v", category.Description(), description)
	}
	if category.Status() != CategoryStatusActive {
		t.Errorf("ItemCategory.Status() = %q, want %q", category.Status(), CategoryStatusActive)
	}
	if category.CreatedAt().Before(before) || category.CreatedAt().After(after) {
		t.Errorf("ItemCategory.CreatedAt() = %v, expected between %v and %v", category.CreatedAt(), before, after)
	}
}

func Test_NewItemCategory_InvalidID(t *testing.T) {
	t.Parallel()

	category, err := NewItemCategory(
		"",
		mustCategoryName(t, "Office Supplies"),
		mustCategoryDescription(t, "General office supplies"),
	)

	if err != ErrInvalidCategoryID {
		t.Fatalf("NewItemCategory() error = %v, want %v", err, ErrInvalidCategoryID)
	}
	if category != nil {
		t.Errorf("NewItemCategory() = %v, want nil", category)
	}
}

func Test_ReconstituteItemCategory(t *testing.T) {
	t.Parallel()

	createdAt := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name        string
		id          string
		value       string
		description string
		status      string
		wantStatus  CategoryStatus
		wantErr     error
	}{
		{
			name:        "valid active category",
			id:          "cat_123",
			value:       "Office Supplies",
			description: "General office supplies",
			status:      "active",
			wantStatus:  CategoryStatusActive,
		},
		{
			name:        "valid inactive category",
			id:          "cat_123",
			value:       "Office Supplies",
			description: "",
			status:      "inactive",
			wantStatus:  CategoryStatusInactive,
		},
		{
			name:        "empty id is allowed",
			id:          "",
			value:       "Office Supplies",
			description: "General office supplies",
			status:      "active",
			wantStatus:  CategoryStatusActive,
		},
		{
			name:        "invalid name",
			id:          "cat_123",
			value:       "",
			description: "General office supplies",
			status:      "active",
			wantErr:     ErrCategoryNameEmpty,
		},
		{
			name:        "invalid description",
			id:          "cat_123",
			value:       "Office Supplies",
			description: strings.Repeat("a", 201),
			status:      "active",
			wantErr:     ErrCategoryDescriptionTooLong,
		},
		{
			name:        "invalid status",
			id:          "cat_123",
			value:       "Office Supplies",
			description: "General office supplies",
			status:      "archived",
			wantErr:     ErrInvalidCategoryStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			category, err := ReconstituteItemCategory(
				tt.id,
				tt.value,
				tt.description,
				tt.status,
				createdAt,
			)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("ReconstituteItemCategory() error = %v, want %v", err, tt.wantErr)
				}
				if category != nil {
					t.Errorf("ReconstituteItemCategory() = %v, want nil", category)
				}
				return
			}

			if err != nil {
				t.Fatalf("ReconstituteItemCategory() unexpected error = %v", err)
			}
			if category == nil {
				t.Fatal("ReconstituteItemCategory() = nil, want category")
			}
			if category.ID() != tt.id {
				t.Errorf("ItemCategory.ID() = %q, want %q", category.ID(), tt.id)
			}
			if category.Name().Value() != tt.value {
				t.Errorf("ItemCategory.Name().Value() = %q, want %q", category.Name().Value(), tt.value)
			}
			if category.Description().Value() != tt.description {
				t.Errorf("ItemCategory.Description().Value() = %q, want %q", category.Description().Value(), tt.description)
			}
			if category.Status() != tt.wantStatus {
				t.Errorf("ItemCategory.Status() = %q, want %q", category.Status(), tt.wantStatus)
			}
			if !category.CreatedAt().Equal(createdAt) {
				t.Errorf("ItemCategory.CreatedAt() = %v, want %v", category.CreatedAt(), createdAt)
			}
		})
	}
}

func Test_ItemCategory_Update(t *testing.T) {
	t.Parallel()

	strPtr := func(value string) *string { return &value }

	tests := []struct {
		name       string
		updateName *string
		updateDesc *string
		wantName   string
		wantDesc   string
		wantErr    error
	}{
		{
			name:       "update name only",
			updateName: strPtr("Hardware"),
			wantName:   "Hardware",
			wantDesc:   "General office supplies",
		},
		{
			name:       "update description only",
			updateDesc: strPtr("Updated description"),
			wantName:   "Office Supplies",
			wantDesc:   "Updated description",
		},
		{
			name:       "empty description is allowed",
			updateDesc: strPtr(""),
			wantName:   "Office Supplies",
			wantDesc:   "",
		},
		{
			name:     "no fields provided",
			wantErr:  ErrNoCategoryUpdateFields,
			wantName: "Office Supplies",
			wantDesc: "General office supplies",
		},
		{
			name:       "invalid name leaves other fields unchanged",
			updateName: strPtr(""),
			updateDesc: strPtr("Ignored description"),
			wantErr:    ErrCategoryNameEmpty,
			wantName:   "Office Supplies",
			wantDesc:   "General office supplies",
		},
		{
			name:       "invalid description leaves other fields unchanged",
			updateName: strPtr("Hardware"),
			updateDesc: strPtr(strings.Repeat("a", 201)),
			wantErr:    ErrCategoryDescriptionTooLong,
			wantName:   "Office Supplies",
			wantDesc:   "General office supplies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			category := mustCategoryWithStatus(t, CategoryStatusActive)

			err := category.Update(tt.updateName, tt.updateDesc)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("ItemCategory.Update() error = %v, want %v", err, tt.wantErr)
				}
				if category.Name().Value() != tt.wantName {
					t.Errorf("ItemCategory.Name().Value() = %q, want %q", category.Name().Value(), tt.wantName)
				}
				if category.Description().Value() != tt.wantDesc {
					t.Errorf("ItemCategory.Description().Value() = %q, want %q", category.Description().Value(), tt.wantDesc)
				}
				return
			}

			if err != nil {
				t.Fatalf("ItemCategory.Update() unexpected error = %v", err)
			}
			if category.Name().Value() != tt.wantName {
				t.Errorf("ItemCategory.Name().Value() = %q, want %q", category.Name().Value(), tt.wantName)
			}
			if category.Description().Value() != tt.wantDesc {
				t.Errorf("ItemCategory.Description().Value() = %q, want %q", category.Description().Value(), tt.wantDesc)
			}
		})
	}
}

func Test_ItemCategory_ActivateAndDeactivate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		initial   CategoryStatus
		activate  bool
		wantState CategoryStatus
	}{
		{
			name:      "active category can be deactivated",
			initial:   CategoryStatusActive,
			activate:  false,
			wantState: CategoryStatusInactive,
		},
		{
			name:      "inactive category can be activated",
			initial:   CategoryStatusInactive,
			activate:  true,
			wantState: CategoryStatusActive,
		},
		{
			name:      "active category stays active when activated",
			initial:   CategoryStatusActive,
			activate:  true,
			wantState: CategoryStatusActive,
		},
		{
			name:      "inactive category stays inactive when deactivated",
			initial:   CategoryStatusInactive,
			activate:  false,
			wantState: CategoryStatusInactive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			category := mustCategoryWithStatus(t, tt.initial)

			if tt.activate {
				category.Activate()
			} else {
				category.Deactivate()
			}

			if category.Status() != tt.wantState {
				t.Errorf("ItemCategory.Status() = %q, want %q", category.Status(), tt.wantState)
			}
		})
	}
}

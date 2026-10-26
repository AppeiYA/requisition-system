package domain

import (
	"testing"
	"time"

	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

func mustCategoryName(t *testing.T, value string) *CategoryName {
	t.Helper()

	name, err := NewCategoryName(value)
	if err != nil {
		t.Fatalf("NewCategoryName(%q) unexpected error = %v", value, err)
	}
	return name
}

func mustCategoryDescription(t *testing.T, value string) *CategoryDescription {
	t.Helper()

	description, err := NewCategoryDescription(value)
	if err != nil {
		t.Fatalf("NewCategoryDescription(%q) unexpected error = %v", value, err)
	}
	return description
}

func mustItemName(t *testing.T, value string) *ItemName {
	t.Helper()

	name, err := NewItemName(value)
	if err != nil {
		t.Fatalf("NewItemName(%q) unexpected error = %v", value, err)
	}
	return name
}

func mustItemDescription(t *testing.T, value string) *ItemDescription {
	t.Helper()

	description, err := NewItemDescription(value)
	if err != nil {
		t.Fatalf("NewItemDescription(%q) unexpected error = %v", value, err)
	}
	return description
}

func mustRequisitionJustification(t *testing.T, value string) *RequisitionJustification {
	t.Helper()

	justification, err := NewRequisitionJustification(value)
	if err != nil {
		t.Fatalf("NewRequisitionJustification(%q) unexpected error = %v", value, err)
	}
	return justification
}

func mustRequisitionQuantity(t *testing.T, quantity int) *RequisitionQuantity {
	t.Helper()

	value, err := NewRequisitionQuantity(quantity)
	if err != nil {
		t.Fatalf("NewRequisitionQuantity(%d) unexpected error = %v", quantity, err)
	}
	return value
}

func mustNairaMoney(t *testing.T, amount float64) *shareddomain.Money {
	t.Helper()

	money, err := shareddomain.NewNairaMoney(amount)
	if err != nil {
		t.Fatalf("NewNairaMoney(%v) unexpected error = %v", amount, err)
	}
	return money
}

func mustCategoryWithStatus(t *testing.T, status CategoryStatus) *ItemCategory {
	t.Helper()

	category, err := ReconstituteItemCategory(
		"cat_123",
		"Office Supplies",
		"General office supplies",
		status.String(),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("ReconstituteItemCategory() unexpected error = %v", err)
	}
	return category
}

func mustItemWithStatus(t *testing.T, status ItemStatus) *Item {
	t.Helper()

	item, err := ReconstituteItem(
		"itm_123",
		"Laptop",
		"Business laptop",
		"cat_123",
		0,
		status.String(),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("ReconstituteItem() unexpected error = %v", err)
	}
	return item
}

func mustRequisitionWithStatus(
	t *testing.T,
	status RequisitionStatus,
	createdAt time.Time,
	updatedAt time.Time,
) *Requisition {
	t.Helper()

	requisition, err := ReconstituteRequisition(
		"req_123",
		"usr_123",
		"itm_123",
		"cat_123",
		"Laptop",
		"Business laptop",
		2,
		mustNairaMoney(t, 1500),
		"Required for the new hire",
		PriorityNormal.String(),
		createdAt.Add(24*time.Hour),
		status.String(),
		createdAt,
		updatedAt,
	)
	if err != nil {
		t.Fatalf("ReconstituteRequisition() unexpected error = %v", err)
	}
	return requisition
}

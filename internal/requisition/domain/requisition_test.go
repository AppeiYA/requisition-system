package domain

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

func Test_NewRequisition(t *testing.T) {
	t.Parallel()

	name := mustItemName(t, "Laptop")
	description := mustItemDescription(t, "Business laptop")
	quantity := mustRequisitionQuantity(t, 2)
	cost := mustNairaMoney(t, 1500)
	justification := mustRequisitionJustification(t, "Required for the new hire")
	neededBy := time.Now().Add(24 * time.Hour)

	before := time.Now()
	requisition, err := NewRequisition(
		"req_123",
		"usr_123",
		"dpt_123",
		"itm_123",
		"cat_123",
		name,
		description,
		quantity,
		cost,
		justification,
		PriorityUrgent,
		neededBy,
	)
	after := time.Now()

	if err != nil {
		t.Fatalf("NewRequisition() unexpected error = %v", err)
	}
	if requisition == nil {
		t.Fatal("NewRequisition() = nil, want requisition")
	}
	if requisition.ID() != "req_123" {
		t.Errorf("Requisition.ID() = %q, want %q", requisition.ID(), "req_123")
	}
	if requisition.RequestedBy() != "usr_123" {
		t.Errorf("Requisition.RequestedBy() = %q, want %q", requisition.RequestedBy(), "usr_123")
	}
	if requisition.DepartmentID() != "dpt_123" {
		t.Errorf("Requisition.DepartmentID() = %q, want %q", requisition.DepartmentID(), "dpt_123")
	}
	if requisition.ItemID() != "itm_123" {
		t.Errorf("Requisition.ItemID() = %q, want %q", requisition.ItemID(), "itm_123")
	}
	if requisition.CategoryID() != "cat_123" {
		t.Errorf("Requisition.CategoryID() = %q, want %q", requisition.CategoryID(), "cat_123")
	}
	if requisition.RequestedItemName() != name {
		t.Errorf("Requisition.RequestedItemName() = %v, want %v", requisition.RequestedItemName(), name)
	}
	if requisition.RequestedItemDescription() != description {
		t.Errorf("Requisition.RequestedItemDescription() = %v, want %v", requisition.RequestedItemDescription(), description)
	}
	if requisition.Quantity() != quantity {
		t.Errorf("Requisition.Quantity() = %v, want %v", requisition.Quantity(), quantity)
	}
	if requisition.EstimatedCost() != cost {
		t.Errorf("Requisition.EstimatedCost() = %v, want %v", requisition.EstimatedCost(), cost)
	}
	if requisition.Justification() != justification {
		t.Errorf("Requisition.Justification() = %v, want %v", requisition.Justification(), justification)
	}
	if requisition.Priority() != PriorityUrgent {
		t.Errorf("Requisition.Priority() = %q, want %q", requisition.Priority(), PriorityUrgent)
	}
	if !requisition.NeededBy().Equal(neededBy) {
		t.Errorf("Requisition.NeededBy() = %v, want %v", requisition.NeededBy(), neededBy)
	}
	if requisition.Status() != RequisitionStatusPending {
		t.Errorf("Requisition.Status() = %q, want %q", requisition.Status(), RequisitionStatusPending)
	}
	if requisition.CreatedAt().Before(before) || requisition.CreatedAt().After(after) {
		t.Errorf("Requisition.CreatedAt() = %v, expected between %v and %v", requisition.CreatedAt(), before, after)
	}
	if requisition.UpdatedAt().Before(before) || requisition.UpdatedAt().After(after) {
		t.Errorf("Requisition.UpdatedAt() = %v, expected between %v and %v", requisition.UpdatedAt(), before, after)
	}
	if !requisition.CreatedAt().Equal(requisition.UpdatedAt()) {
		t.Errorf("Requisition.CreatedAt() = %v, want UpdatedAt() = %v", requisition.CreatedAt(), requisition.UpdatedAt())
	}
}

func Test_NewRequisition_AllowsEmptyItemID(t *testing.T) {
	t.Parallel()

	requisition, err := NewRequisition(
		"req_123",
		"usr_123",
		"dpt_123",
		"",
		"cat_123",
		mustItemName(t, "Laptop"),
		mustItemDescription(t, "Business laptop"),
		mustRequisitionQuantity(t, 1),
		mustNairaMoney(t, 100),
		mustRequisitionJustification(t, "Required for the new hire"),
		PriorityNormal,
		time.Now().Add(time.Hour),
	)

	if err != nil {
		t.Fatalf("NewRequisition() unexpected error = %v", err)
	}
	if requisition.ItemID() != "" {
		t.Errorf("Requisition.ItemID() = %q, want empty", requisition.ItemID())
	}
	if requisition.CategoryID() != "cat_123" {
		t.Errorf("Requisition.CategoryID() = %q, want %q", requisition.CategoryID(), "cat_123")
	}
}

func Test_NewRequisition_Validation(t *testing.T) {
	t.Parallel()

	type input struct {
		id            string
		requestedBy   string
		departmentID  string
		itemID        string
		categoryID    string
		name          *ItemName
		description   *ItemDescription
		quantity      *RequisitionQuantity
		cost          *shareddomain.Money
		justification *RequisitionJustification
		priority      RequisitionPriority
		neededBy      time.Time
	}

	base := func() input {
		return input{
			id:            "req_123",
			requestedBy:   "usr_123",
			departmentID:  "dpt_123",
			itemID:        "itm_123",
			categoryID:    "cat_123",
			name:          mustItemName(t, "Laptop"),
			description:   mustItemDescription(t, "Business laptop"),
			quantity:      mustRequisitionQuantity(t, 2),
			cost:          mustNairaMoney(t, 1500),
			justification: mustRequisitionJustification(t, "Required for the new hire"),
			priority:      PriorityNormal,
			neededBy:      time.Now().Add(time.Hour),
		}
	}

	tests := []struct {
		name       string
		mutate     func(*input)
		wantErrMsg string
	}{
		{
			name:       "empty id",
			mutate:     func(in *input) { in.id = "" },
			wantErrMsg: ErrInvalidRequisitionID.Error(),
		},
		{
			name:       "invalid id prefix",
			mutate:     func(in *input) { in.id = "REQ_123" },
			wantErrMsg: ErrInvalidRequisitionID.Error(),
		},
		{
			name:       "empty requestedBy",
			mutate:     func(in *input) { in.requestedBy = "" },
			wantErrMsg: ErrRequestedIDCannotBeEmpty.Error(),
		},
		{
			name:       "empty departmentID",
			mutate:     func(in *input) { in.departmentID = "" },
			wantErrMsg: ErrDepartmentIDCannotBeEmpty.Error(),
		},
		{
			name:       "invalid department id prefix",
			mutate:     func(in *input) { in.departmentID = "dept_123" },
			wantErrMsg: ErrInvalidDepartmentID.Error(),
		},
		{
			name:       "empty category id",
			mutate:     func(in *input) { in.categoryID = "" },
			wantErrMsg: ErrInvalidCategoryID.Error(),
		},
		{
			name:       "invalid category id prefix",
			mutate:     func(in *input) { in.categoryID = "CAT_123" },
			wantErrMsg: ErrInvalidCategoryID.Error(),
		},
		{
			name:       "invalid item id prefix",
			mutate:     func(in *input) { in.itemID = "ITEM_123" },
			wantErrMsg: ErrInvalidItemID.Error(),
		},
		{
			name:       "nil requested item name",
			mutate:     func(in *input) { in.name = nil },
			wantErrMsg: ErrNilFields("requestedItemName").Error(),
		},
		{
			name:       "nil requested item description",
			mutate:     func(in *input) { in.description = nil },
			wantErrMsg: ErrNilFields("requestedItemDescription").Error(),
		},
		{
			name:       "nil justification",
			mutate:     func(in *input) { in.justification = nil },
			wantErrMsg: ErrNilFields("justification").Error(),
		},
		{
			name:       "nil quantity",
			mutate:     func(in *input) { in.quantity = nil },
			wantErrMsg: ErrNilFields("quantity").Error(),
		},
		{
			name:       "nil estimated cost",
			mutate:     func(in *input) { in.cost = nil },
			wantErrMsg: ErrNilFields("estimatedCost").Error(),
		},
		{
			name:       "neededBy date in the past",
			mutate:     func(in *input) { in.neededBy = time.Now().AddDate(0, 0, -1) },
			wantErrMsg: ErrInvalidNeededByDate.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			in := base()
			tt.mutate(&in)

			requisition, err := NewRequisition(
				in.id,
				in.requestedBy,
				in.departmentID,
				in.itemID,
				in.categoryID,
				in.name,
				in.description,
				in.quantity,
				in.cost,
				in.justification,
				in.priority,
				in.neededBy,
			)

			if err == nil {
				t.Fatalf("NewRequisition() error = nil, want %q", tt.wantErrMsg)
			}
			if err.Error() != tt.wantErrMsg {
				t.Errorf("NewRequisition() error = %q, want %q", err.Error(), tt.wantErrMsg)
			}
			if requisition != nil {
				t.Errorf("NewRequisition() = %v, want nil", requisition)
			}
		})
	}
}

func Test_ReconstituteRequisition(t *testing.T) {
	t.Parallel()

	createdAt := time.Now().Add(-48 * time.Hour)
	updatedAt := createdAt.Add(2 * time.Hour)
	neededBy := time.Now().Add(24 * time.Hour)

	type input struct {
		id                       string
		requestedBy              string
		departmentID             string
		itemID                   string
		categoryID               string
		requestedItemName        string
		requestedItemDescription string
		quantity                 int
		estimatedCost            *shareddomain.Money
		justification            string
		priority                 string
		neededBy                 time.Time
		status                   string
		createdAt                time.Time
		updatedAt                time.Time
	}

	base := func() input {
		return input{
			id:                       "req_123",
			requestedBy:              "usr_123",
			departmentID:             "dpt_123",
			itemID:                   "itm_123",
			categoryID:               "cat_123",
			requestedItemName:        "Laptop",
			requestedItemDescription: "Business laptop",
			quantity:                 2,
			estimatedCost:            mustNairaMoney(t, 1500),
			justification:            "Required for the new hire",
			priority:                 "normal",
			neededBy:                 neededBy,
			status:                   "approved",
			createdAt:                createdAt,
			updatedAt:                updatedAt,
		}
	}

	tests := []struct {
		name         string
		mutate       func(*input)
		wantErr      error
		wantStatus   RequisitionStatus
		wantPriority RequisitionPriority
	}{
		{
			name:         "valid requisition",
			wantStatus:   RequisitionStatusApproved,
			wantPriority: PriorityNormal,
		},
		{
			name: "valid requisition with empty item ID",
			mutate: func(in *input) {
				in.itemID = ""
				in.priority = "low"
				in.status = "pending"
			},
			wantStatus:   RequisitionStatusPending,
			wantPriority: PriorityLow,
		},
		{
			name:    "empty id",
			mutate:  func(in *input) { in.id = "" },
			wantErr: ErrInvalidRequisitionID,
		},
		{
			name:    "invalid id prefix",
			mutate:  func(in *input) { in.id = "REQ_123" },
			wantErr: ErrInvalidRequisitionID,
		},
		{
			name:    "empty requestedBy",
			mutate:  func(in *input) { in.requestedBy = "" },
			wantErr: ErrRequestedIDCannotBeEmpty,
		},
		{
			name:    "empty departmentID",
			mutate:  func(in *input) { in.departmentID = "" },
			wantErr: ErrDepartmentIDCannotBeEmpty,
		},
		{
			name:    "invalid department id prefix",
			mutate:  func(in *input) { in.departmentID = "dept_123" },
			wantErr: ErrInvalidDepartmentID,
		},
		{
			name:    "invalid category id prefix",
			mutate:  func(in *input) { in.categoryID = "CAT_123" },
			wantErr: ErrInvalidCategoryID,
		},
		{
			name:    "invalid item id prefix",
			mutate:  func(in *input) { in.itemID = "ITEM_123" },
			wantErr: ErrInvalidItemID,
		},
		{
			name:    "invalid item name",
			mutate:  func(in *input) { in.requestedItemName = "" },
			wantErr: ErrItemNameCannotBeEmpty,
		},
		{
			name:    "item name too long",
			mutate:  func(in *input) { in.requestedItemName = strings.Repeat("a", 101) },
			wantErr: ErrItemNameTooLong,
		},
		{
			name:    "invalid item description",
			mutate:  func(in *input) { in.requestedItemDescription = "" },
			wantErr: ErrItemDescriptionCannotBeEmpty,
		},
		{
			name:    "item description too long",
			mutate:  func(in *input) { in.requestedItemDescription = strings.Repeat("a", 201) },
			wantErr: ErrItemDescriptionTooLong,
		},
		{
			name:    "invalid quantity",
			mutate:  func(in *input) { in.quantity = 0 },
			wantErr: ErrInvalidRequisitionQuantity,
		},
		{
			name:    "quantity exceeds limit",
			mutate:  func(in *input) { in.quantity = 1001 },
			wantErr: ErrRequisitionQuantityExceedsLimit,
		},
		{
			name:    "empty justification",
			mutate:  func(in *input) { in.justification = "" },
			wantErr: ErrInvalidRequisitionJustification,
		},
		{
			name:    "justification too long",
			mutate:  func(in *input) { in.justification = strings.Repeat("a", 501) },
			wantErr: ErrRequisitionJustificationTooLong,
		},
		{
			name:    "invalid priority",
			mutate:  func(in *input) { in.priority = "critical" },
			wantErr: ErrInvalidRequisitionPriority,
		},
		{
			name:    "invalid status",
			mutate:  func(in *input) { in.status = "draft" },
			wantErr: ErrInvalidRequisitionStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			in := base()
			if tt.mutate != nil {
				tt.mutate(&in)
			}

			requisition, err := ReconstituteRequisition(
				in.id,
				in.requestedBy,
				in.departmentID,
				in.itemID,
				in.categoryID,
				in.requestedItemName,
				in.requestedItemDescription,
				in.quantity,
				in.estimatedCost,
				in.justification,
				in.priority,
				in.neededBy,
				in.status,
				in.createdAt,
				in.updatedAt,
			)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("ReconstituteRequisition() error = %v, want %v", err, tt.wantErr)
				}
				if requisition != nil {
					t.Errorf("ReconstituteRequisition() = %v, want nil", requisition)
				}
				return
			}

			if err != nil {
				t.Fatalf("ReconstituteRequisition() unexpected error = %v", err)
			}
			if requisition == nil {
				t.Fatal("ReconstituteRequisition() = nil, want requisition")
			}
			if requisition.ID() != in.id {
				t.Errorf("Requisition.ID() = %q, want %q", requisition.ID(), in.id)
			}
			if requisition.RequestedBy() != in.requestedBy {
				t.Errorf("Requisition.RequestedBy() = %q, want %q", requisition.RequestedBy(), in.requestedBy)
			}
			if requisition.DepartmentID() != in.departmentID {
				t.Errorf("Requisition.DepartmentID() = %q, want %q", requisition.DepartmentID(), in.departmentID)
			}
			if requisition.ItemID() != in.itemID {
				t.Errorf("Requisition.ItemID() = %q, want %q", requisition.ItemID(), in.itemID)
			}
			if requisition.CategoryID() != in.categoryID {
				t.Errorf("Requisition.CategoryID() = %q, want %q", requisition.CategoryID(), in.categoryID)
			}
			if requisition.RequestedItemName().Value() != in.requestedItemName {
				t.Errorf("Requisition.RequestedItemName().Value() = %q, want %q", requisition.RequestedItemName().Value(), in.requestedItemName)
			}
			if requisition.RequestedItemDescription().Value() != in.requestedItemDescription {
				t.Errorf("Requisition.RequestedItemDescription().Value() = %q, want %q", requisition.RequestedItemDescription().Value(), in.requestedItemDescription)
			}
			if requisition.Quantity().Value() != int64(in.quantity) {
				t.Errorf("Requisition.Quantity().Value() = %d, want %d", requisition.Quantity().Value(), in.quantity)
			}
			if requisition.EstimatedCost() != in.estimatedCost {
				t.Errorf("Requisition.EstimatedCost() = %v, want %v", requisition.EstimatedCost(), in.estimatedCost)
			}
			if requisition.Justification().Value() != in.justification {
				t.Errorf("Requisition.Justification().Value() = %q, want %q", requisition.Justification().Value(), in.justification)
			}
			if requisition.Priority() != tt.wantPriority {
				t.Errorf("Requisition.Priority() = %q, want %q", requisition.Priority(), tt.wantPriority)
			}
			if !requisition.NeededBy().Equal(in.neededBy) {
				t.Errorf("Requisition.NeededBy() = %v, want %v", requisition.NeededBy(), in.neededBy)
			}
			if requisition.Status() != tt.wantStatus {
				t.Errorf("Requisition.Status() = %q, want %q", requisition.Status(), tt.wantStatus)
			}
			if !requisition.CreatedAt().Equal(in.createdAt) {
				t.Errorf("Requisition.CreatedAt() = %v, want %v", requisition.CreatedAt(), in.createdAt)
			}
			if !requisition.UpdatedAt().Equal(in.updatedAt) {
				t.Errorf("Requisition.UpdatedAt() = %v, want %v", requisition.UpdatedAt(), in.updatedAt)
			}
		})
	}
}

func Test_Requisition_UpdateStatus(t *testing.T) {
	t.Parallel()

	allowedTransitions := map[RequisitionStatus][]RequisitionStatus{
		RequisitionStatusPending: {
			RequisitionStatusApproved,
			RequisitionStatusRejected,
			RequisitionStatusCancelled,
		},
		RequisitionStatusApproved: {
			RequisitionStatusInFulfillment,
			RequisitionStatusCancelled,
		},
		RequisitionStatusInFulfillment: {
			RequisitionStatusCompleted,
			RequisitionStatusCancelled,
		},
	}

	statuses := []RequisitionStatus{
		RequisitionStatusPending,
		RequisitionStatusApproved,
		RequisitionStatusRejected,
		RequisitionStatusCompleted,
		RequisitionStatusCancelled,
		RequisitionStatusInFulfillment,
	}

	for _, from := range statuses {
		for _, to := range statuses {
			t.Run(fmt.Sprintf("%s_to_%s", from, to), func(t *testing.T) {
				t.Parallel()

				createdAt := time.Now().Add(-48 * time.Hour)
				updatedAt := createdAt.Add(2 * time.Hour)
				requisition := mustRequisitionWithStatus(t, from, createdAt, updatedAt)

				before := time.Now()
				changed := requisition.UpdateStatus(to)
				after := time.Now()

				want := slices.Contains(allowedTransitions[from], to)

				if changed != want {
					t.Fatalf("Requisition.UpdateStatus() = %v, want %v", changed, want)
				}
				if !requisition.CreatedAt().Equal(createdAt) {
					t.Errorf("Requisition.CreatedAt() = %v, want %v", requisition.CreatedAt(), createdAt)
				}

				if !want {
					if requisition.Status() != from {
						t.Errorf("Requisition.Status() = %q, want unchanged %q", requisition.Status(), from)
					}
					if !requisition.UpdatedAt().Equal(updatedAt) {
						t.Errorf("Requisition.UpdatedAt() = %v, want unchanged %v", requisition.UpdatedAt(), updatedAt)
					}
					return
				}

				if requisition.Status() != to {
					t.Errorf("Requisition.Status() = %q, want %q", requisition.Status(), to)
				}
				if requisition.UpdatedAt().Before(before) || requisition.UpdatedAt().After(after) {
					t.Errorf("Requisition.UpdatedAt() = %v, expected between %v and %v", requisition.UpdatedAt(), before, after)
				}
			})
		}
	}
}

func Test_Requisition_UpdateStatus_UnknownStatus(t *testing.T) {
	t.Parallel()

	now := time.Now()
	requisition := &Requisition{
		status:    RequisitionStatus("draft"),
		createdAt: now,
		updatedAt: now,
	}

	if requisition.UpdateStatus(RequisitionStatusPending) {
		t.Error("Requisition.UpdateStatus() = true, want false")
	}
	if requisition.Status() != RequisitionStatus("draft") {
		t.Errorf("Requisition.Status() = %q, want %q", requisition.Status(), RequisitionStatus("draft"))
	}
	if !requisition.UpdatedAt().Equal(now) {
		t.Errorf("Requisition.UpdatedAt() = %v, want %v", requisition.UpdatedAt(), now)
	}
}

func Test_Requisition_UpdateRequisition(t *testing.T) {
	t.Parallel()

	t.Run("successful update with all fields", func(t *testing.T) {
		t.Parallel()
		now := time.Now().Add(-1 * time.Hour)
		req := mustRequisitionWithStatus(t, RequisitionStatusPending, now, now)

		newQty := mustRequisitionQuantity(t, 5)
		newCost := mustNairaMoney(t, 3000)
		newJust := mustRequisitionJustification(t, "Updated justification")
		newPriority := PriorityUrgent
		newNeededBy := time.Now().Add(48 * time.Hour)

		before := time.Now()
		err := req.UpdateRequisition(&UpdateRequisitionInput{
			Quantity:      newQty,
			EstimatedCost: newCost,
			Justification: newJust,
			Priority:      &newPriority,
			NeededBy:      &newNeededBy,
		})
		after := time.Now()

		if err != nil {
			t.Fatalf("UpdateRequisition() unexpected error: %v", err)
		}
		if req.Quantity() != newQty {
			t.Errorf("Quantity() = %v, want %v", req.Quantity(), newQty)
		}
		if req.EstimatedCost() != newCost {
			t.Errorf("EstimatedCost() = %v, want %v", req.EstimatedCost(), newCost)
		}
		if req.Justification() != newJust {
			t.Errorf("Justification() = %v, want %v", req.Justification(), newJust)
		}
		if req.Priority() != PriorityUrgent {
			t.Errorf("Priority() = %v, want %v", req.Priority(), PriorityUrgent)
		}
		if !req.NeededBy().Equal(newNeededBy) {
			t.Errorf("NeededBy() = %v, want %v", req.NeededBy(), newNeededBy)
		}
		if req.UpdatedAt().Before(before) || req.UpdatedAt().After(after) {
			t.Errorf("UpdatedAt() = %v, expected between %v and %v", req.UpdatedAt(), before, after)
		}
	})

	t.Run("partial update without priority or neededBy", func(t *testing.T) {
		t.Parallel()
		now := time.Now().Add(-1 * time.Hour)
		req := mustRequisitionWithStatus(t, RequisitionStatusPending, now, now)
		originalPriority := req.Priority()
		originalNeededBy := req.NeededBy()

		newQty := mustRequisitionQuantity(t, 10)
		err := req.UpdateRequisition(&UpdateRequisitionInput{
			Quantity: newQty,
		})

		if err != nil {
			t.Fatalf("UpdateRequisition() unexpected error: %v", err)
		}
		if req.Quantity() != newQty {
			t.Errorf("Quantity() = %v, want %v", req.Quantity(), newQty)
		}
		if req.Priority() != originalPriority {
			t.Errorf("Priority() should remain %v, got %v", originalPriority, req.Priority())
		}
		if !req.NeededBy().Equal(originalNeededBy) {
			t.Errorf("NeededBy() should remain %v, got %v", originalNeededBy, req.NeededBy())
		}
	})

	t.Run("fails when updating neededBy to past date", func(t *testing.T) {
		t.Parallel()
		req := mustRequisitionWithStatus(t, RequisitionStatusPending, time.Now(), time.Now())
		past := time.Now().AddDate(0, 0, -1)

		err := req.UpdateRequisition(&UpdateRequisitionInput{
			NeededBy: &past,
		})
		if err == nil {
			t.Fatal("expected error for past neededBy date, got nil")
		}
	})

	t.Run("fails when updating with invalid priority", func(t *testing.T) {
		t.Parallel()
		req := mustRequisitionWithStatus(t, RequisitionStatusPending, time.Now(), time.Now())
		invalidPriority := RequisitionPriority("invalid")

		err := req.UpdateRequisition(&UpdateRequisitionInput{
			Priority: &invalidPriority,
		})
		if err == nil {
			t.Fatal("expected error for invalid priority, got nil")
		}
	})

	t.Run("fails when requisition is not in pending status", func(t *testing.T) {
		t.Parallel()
		req := mustRequisitionWithStatus(t, RequisitionStatusApproved, time.Now(), time.Now())

		err := req.UpdateRequisition(&UpdateRequisitionInput{
			Quantity: mustRequisitionQuantity(t, 5),
		})
		if err == nil {
			t.Fatal("expected error when updating approved requisition, got nil")
		}
	})

	t.Run("fails with nil input", func(t *testing.T) {
		t.Parallel()
		req := mustRequisitionWithStatus(t, RequisitionStatusPending, time.Now(), time.Now())

		err := req.UpdateRequisition(nil)
		if err == nil {
			t.Fatal("expected error for nil input, got nil")
		}
	})
}

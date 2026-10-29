package domain

import (
	"testing"
	"time"
)

func aggregateItem(t *testing.T, id, requestedID string, line int) *RequisitionItem {
	t.Helper()
	item, err := NewRequisitionItem(
		"reqitem_"+string(rune('a'+line)),
		"req_123",
		line,
		id,
		requestedID,
		mustItemName(t, "Laptop"),
		mustRequisitionQuantity(t, 2),
		mustNairaMoney(t, 1500),
		"for the new team member",
	)
	if err != nil {
		t.Fatalf("NewRequisitionItem() unexpected error = %v", err)
	}
	return item
}

func validSubmittedRequisition(t *testing.T) *Requisition {
	t.Helper()
	requisition, err := NewSubmittedRequisition(
		"req_123",
		"REQ-2026-10001",
		"usr_123",
		"dpt_123",
		PriorityMedium,
		time.Now().UTC().AddDate(0, 0, 1),
		mustRequisitionJustification(t, "Required for the new team member"),
		[]*RequisitionItem{aggregateItem(t, "itm_123", "", 1)},
	)
	if err != nil {
		t.Fatalf("NewSubmittedRequisition() unexpected error = %v", err)
	}
	return requisition
}

func TestNewRequisitionItemRequiresExactlyOneSelection(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		itemID      string
		requestedID string
	}{
		{name: "both selected", itemID: "itm_123", requestedID: "ritem_123"},
		{name: "neither selected"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewRequisitionItem(
				"reqitem_1", "req_123", 1, tt.itemID, tt.requestedID,
				mustItemName(t, "Laptop"), mustRequisitionQuantity(t, 1),
				mustNairaMoney(t, 0), "",
			)
			if err != ErrInvalidItemSelection {
				t.Fatalf("NewRequisitionItem() error = %v, want %v", err, ErrInvalidItemSelection)
			}
		})
	}
}

func TestNewSubmittedRequisitionValidatesLinesAndTotals(t *testing.T) {
	t.Parallel()

	first := aggregateItem(t, "itm_123", "", 1)
	second := aggregateItem(t, "", "ritem_123", 2)
	requisition, err := NewSubmittedRequisition(
		"req_123", "REQ-2026-10001", "usr_123", "dpt_123",
		PriorityMedium, time.Now().UTC().AddDate(0, 0, 1),
		mustRequisitionJustification(t, "Office supplies"),
		[]*RequisitionItem{first, second},
	)
	if err != nil {
		t.Fatalf("NewSubmittedRequisition() unexpected error = %v", err)
	}
	if len(requisition.Items()) != 2 {
		t.Fatalf("Items() length = %d, want 2", len(requisition.Items()))
	}
	total, err := requisition.TotalEstimatedAmount()
	if err != nil {
		t.Fatalf("TotalEstimatedAmount() unexpected error = %v", err)
	}
	if total.ToFloat64() != 6000 {
		t.Errorf("TotalEstimatedAmount() = %v, want 6000", total.ToFloat64())
	}
	if !requisition.IsEditable() {
		t.Error("new requisition should be editable while pending")
	}
}

func TestNewSubmittedRequisitionRejectsEmptyOrNonFuture(t *testing.T) {
	t.Parallel()

	_, err := NewSubmittedRequisition(
		"req_123", "REQ-2026-10001", "usr_123", "dpt_123",
		PriorityMedium, time.Now().UTC().AddDate(0, 0, 1),
		mustRequisitionJustification(t, "Office supplies"), nil,
	)
	if err != ErrRequisitionNeedsItem {
		t.Fatalf("empty items error = %v, want %v", err, ErrRequisitionNeedsItem)
	}

	_, err = NewSubmittedRequisition(
		"req_123", "REQ-2026-10001", "usr_123", "dpt_123",
		PriorityMedium, time.Now().UTC(),
		mustRequisitionJustification(t, "Office supplies"),
		[]*RequisitionItem{aggregateItem(t, "itm_123", "", 1)},
	)
	if err != ErrInvalidNeededByDate {
		t.Fatalf("today needed-by error = %v, want %v", err, ErrInvalidNeededByDate)
	}
}

func TestRequisitionStatusTransitionCreatesHistory(t *testing.T) {
	t.Parallel()

	requisition := validSubmittedRequisition(t)
	history, err := requisition.ApplyStatusTransition(
		"rsh_1", 1, RequisitionStatusRejected, "usr_123",
		ActorRoleStaff, "No budget available", time.Now(),
	)
	if err != nil {
		t.Fatalf("ApplyStatusTransition() unexpected error = %v", err)
	}
	if requisition.Status() != RequisitionStatusRejected {
		t.Fatalf("Status() = %q, want rejected", requisition.Status())
	}
	if len(requisition.StatusHistory()) != 1 || requisition.StatusHistory()[0] != history {
		t.Fatal("status history was not appended")
	}
	if history.FromStatus() == nil || *history.FromStatus() != RequisitionStatusPending {
		t.Fatal("history should record pending as the previous status")
	}
	if _, err := requisition.ApplyStatusTransition(
		"rsh_2", 2, RequisitionStatusApproved, "usr_123",
		ActorRoleStaff, "", time.Now(),
	); err == nil {
		t.Fatal("expected terminal requisition transition to fail")
	}
	if requisition.IsEditable() {
		t.Error("rejected requisition should not be editable")
	}
}

func TestRequisitionStatusHistoryRequiresReason(t *testing.T) {
	t.Parallel()

	_, err := NewRequisitionStatusHistory(
		"rsh_1", "req_123", 1, nil, RequisitionStatusRejected,
		"usr_123", ActorRoleStaff, "", time.Now(),
	)
	if err != ErrStatusReasonRequired {
		t.Fatalf("NewRequisitionStatusHistory() error = %v, want %v", err, ErrStatusReasonRequired)
	}
}

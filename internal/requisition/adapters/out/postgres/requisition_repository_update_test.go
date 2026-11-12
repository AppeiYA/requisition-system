package requisitionpostgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	shareddb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
)

func updateRepositoryRequisition(t *testing.T) *requisitiondomain.Requisition {
	t.Helper()

	createdAt := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	justification, err := requisitiondomain.NewRequisitionJustification("Office setup")
	if err != nil {
		t.Fatal(err)
	}
	priority, err := requisitiondomain.NewRequisitionPriority("medium")
	if err != nil {
		t.Fatal(err)
	}

	line1 := updateRepositoryLine(t, "reqitem-1", 1, "", "ritem-1", "Laptop stand", 2, 100)
	line2 := updateRepositoryLine(t, "reqitem-2", 2, "itm-2", "", "Monitor", 1, 200)
	line3 := updateRepositoryLine(t, "reqitem-3", 3, "", "ritem-2", "Keyboard", 3, 300)

	requested1, err := requisitiondomain.ReconstituteRequestedItem(
		"ritem-1", "Laptop stand", "cat-1", "Adjustable stand", "usr-1",
		"promoted", "itm-999", createdAt,
	)
	if err != nil {
		t.Fatal(err)
	}
	requested2, err := requisitiondomain.NewRequestedItem(
		"ritem-2", mustRepositoryItemName(t, "Keyboard"), "cat-1",
		mustRepositoryItemDescription(t, "Mechanical keyboard"), "usr-1",
	)
	if err != nil {
		t.Fatal(err)
	}

	requisition, err := requisitiondomain.NewSubmittedRequisitionWithRequestedItems(
		"req-1", "REQ-2026-10001", "usr-1", "dpt-1", priority,
		time.Now().UTC().AddDate(0, 0, 5), justification,
		[]*requisitiondomain.RequisitionItem{line1, line2, line3},
		[]*requisitiondomain.RequestedItem{requested1, requested2},
	)
	if err != nil {
		t.Fatal(err)
	}
	return requisition
}

func updateRepositoryLine(
	t *testing.T,
	id string,
	lineNumber int,
	itemID string,
	requestedItemID string,
	name string,
	quantity int,
	amount int64,
) *requisitiondomain.RequisitionItem {
	t.Helper()

	validQuantity, err := requisitiondomain.NewRequisitionQuantity(quantity)
	if err != nil {
		t.Fatal(err)
	}
	money, err := shareddomain.NewNairaMoneyFromMinorUnits(amount)
	if err != nil {
		t.Fatal(err)
	}
	line, err := requisitiondomain.NewRequisitionItem(
		id, "req-1", lineNumber, itemID, requestedItemID,
		mustRepositoryItemName(t, name), validQuantity, money, "",
	)
	if err != nil {
		t.Fatal(err)
	}
	return line
}

func mustRepositoryItemName(t *testing.T, value string) *requisitiondomain.ItemName {
	t.Helper()
	name, err := requisitiondomain.NewItemName(value)
	if err != nil {
		t.Fatal(err)
	}
	return name
}

func mustRepositoryItemDescription(t *testing.T, value string) *requisitiondomain.ItemDescription {
	t.Helper()
	description, err := requisitiondomain.NewItemDescription(value)
	if err != nil {
		t.Fatal(err)
	}
	return description
}

func TestUpdateRequisitionAggregateReconcilesIdentities(t *testing.T) {
	t.Parallel()

	database, mock := newRequisitionMockDB(t)
	requisition := updateRepositoryRequisition(t)

	mock.ExpectQuery("SELECT id, line_number, item_id, requested_item_id").
		WillReturnRows(sqlmock.NewRows([]string{"id", "line_number", "item_id", "requested_item_id"}).
			AddRow("reqitem-1", 1, nil, "ritem-1").
			AddRow("reqitem-2", 2, nil, "ritem-old").
			AddRow("reqitem-4", 4, nil, "ritem-remove"))

	mock.ExpectExec("UPDATE requisitions").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`(?s)UPDATE requested_items\s+SET name = \$1, category_id = \$2, description = \$3\s+WHERE id = \$4`).
		WithArgs("Laptop stand", "cat-1", "Adjustable stand", "ritem-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("INSERT INTO requested_items").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`(?s)DELETE FROM requisition_items\s+WHERE requisition_id = \$1 AND id = ANY\(\$2\)`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE requisition_items").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE requisition_items").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO requisition_items").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.MatchExpectationsInOrder(false)
	guard := `(?s)DELETE FROM requested_items\s+WHERE id = \$1\s+AND review_status = 'unreviewed'\s+AND promoted_item_id IS NULL\s+AND NOT EXISTS`
	mock.ExpectExec(guard).WithArgs("ritem-old").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(guard).WithArgs("ritem-remove").WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewRequisitionRepository(shareddb.NewRepository(database))
	if err := repo.UpdateRequisitionAggregate(context.Background(), requisition); err != nil {
		t.Fatalf("UpdateRequisitionAggregate() unexpected error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateRequisitionAggregateOnlyUpdatesEditableRequestedItemFields(t *testing.T) {
	t.Parallel()

	database, mock := newRequisitionMockDB(t)
	requisition := updateRepositoryRequisition(t)

	mock.ExpectQuery("SELECT id, line_number, item_id, requested_item_id").
		WillReturnRows(sqlmock.NewRows([]string{"id", "line_number", "item_id", "requested_item_id"}).
			AddRow("reqitem-1", 1, nil, "ritem-1").
			AddRow("reqitem-2", 2, "itm-2", nil).
			AddRow("reqitem-3", 3, "itm-3", nil))

	mock.ExpectExec("UPDATE requisitions").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)UPDATE requested_items\s+SET name = \$1, category_id = \$2, description = \$3\s+WHERE id = \$4`).
		WithArgs("Laptop stand", "cat-1", "Adjustable stand", "ritem-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO requested_items").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE requisition_items").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE requisition_items").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE requisition_items").WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewRequisitionRepository(shareddb.NewRepository(database))
	if err := repo.UpdateRequisitionAggregate(context.Background(), requisition); err != nil {
		t.Fatalf("UpdateRequisitionAggregate() unexpected error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateRequisitionAggregateBreaksLineNumberAndKeyCycles(t *testing.T) {
	t.Parallel()

	database, mock := newRequisitionMockDB(t)
	justification, err := requisitiondomain.NewRequisitionJustification("Office setup")
	if err != nil {
		t.Fatal(err)
	}
	priority, err := requisitiondomain.NewRequisitionPriority("medium")
	if err != nil {
		t.Fatal(err)
	}

	// The two lines keep their identities but swap catalogue items and line
	// numbers, which creates a cycle in every unique constraint.
	line1 := updateRepositoryLine(t, "reqitem-1", 2, "itm-2", "", "Item two", 1, 100)
	line2 := updateRepositoryLine(t, "reqitem-2", 1, "itm-1", "", "Item one", 1, 200)
	requisition, err := requisitiondomain.NewSubmittedRequisition(
		"req-1", "REQ-2026-10001", "usr-1", "dpt-1", priority,
		time.Now().UTC().AddDate(0, 0, 5), justification,
		[]*requisitiondomain.RequisitionItem{line1, line2},
	)
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery("SELECT id, line_number, item_id, requested_item_id").
		WillReturnRows(sqlmock.NewRows([]string{"id", "line_number", "item_id", "requested_item_id"}).
			AddRow("reqitem-1", 1, "itm-1", nil).
			AddRow("reqitem-2", 2, "itm-2", nil))

	mock.ExpectExec("UPDATE requisitions").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)DELETE FROM requisition_items\s+WHERE id = \$1 AND requisition_id = \$2`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE requisition_items").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO requisition_items").WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewRequisitionRepository(shareddb.NewRepository(database))
	if err := repo.UpdateRequisitionAggregate(context.Background(), requisition); err != nil {
		t.Fatalf("UpdateRequisitionAggregate() unexpected error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateRequisitionAggregateHandlesSmallIntBoundaryLineNumbers(t *testing.T) {
	t.Parallel()

	database, mock := newRequisitionMockDB(t)
	justification, err := requisitiondomain.NewRequisitionJustification("Office setup")
	if err != nil {
		t.Fatal(err)
	}
	priority, err := requisitiondomain.NewRequisitionPriority("medium")
	if err != nil {
		t.Fatal(err)
	}

	// Swap line 1 and line 32767. The old offset-based staging could not move
	// the persisted 32767 line without overflowing SMALLINT.
	line1 := updateRepositoryLine(t, "reqitem-1", 32767, "itm-2", "", "Item two", 1, 100)
	line2 := updateRepositoryLine(t, "reqitem-2", 1, "itm-1", "", "Item one", 1, 200)
	requisition, err := requisitiondomain.NewSubmittedRequisition(
		"req-1", "REQ-2026-10001", "usr-1", "dpt-1", priority,
		time.Now().UTC().AddDate(0, 0, 5), justification,
		[]*requisitiondomain.RequisitionItem{line1, line2},
	)
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery("SELECT id, line_number, item_id, requested_item_id").
		WillReturnRows(sqlmock.NewRows([]string{"id", "line_number", "item_id", "requested_item_id"}).
			AddRow("reqitem-1", 1, "itm-1", nil).
			AddRow("reqitem-2", 32767, "itm-2", nil))

	mock.ExpectExec("UPDATE requisitions").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)DELETE FROM requisition_items\s+WHERE id = \$1 AND requisition_id = \$2`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE requisition_items").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO requisition_items").WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewRequisitionRepository(shareddb.NewRepository(database))
	if err := repo.UpdateRequisitionAggregate(context.Background(), requisition); err != nil {
		t.Fatalf("UpdateRequisitionAggregate() unexpected error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateRequisitionAggregateRejectsLineNumbersOutsideSmallInt(t *testing.T) {
	t.Parallel()

	database, mock := newRequisitionMockDB(t)
	justification, err := requisitiondomain.NewRequisitionJustification("Office setup")
	if err != nil {
		t.Fatal(err)
	}
	priority, err := requisitiondomain.NewRequisitionPriority("medium")
	if err != nil {
		t.Fatal(err)
	}

	line := updateRepositoryLine(t, "reqitem-1", 32768, "itm-1", "", "Item", 1, 100)
	requisition, err := requisitiondomain.NewSubmittedRequisition(
		"req-1", "REQ-2026-10001", "usr-1", "dpt-1", priority,
		time.Now().UTC().AddDate(0, 0, 5), justification,
		[]*requisitiondomain.RequisitionItem{line},
	)
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery("SELECT id, line_number, item_id, requested_item_id").
		WillReturnRows(sqlmock.NewRows([]string{"id", "line_number", "item_id", "requested_item_id"}))

	repo := NewRequisitionRepository(shareddb.NewRepository(database))
	err = repo.UpdateRequisitionAggregate(context.Background(), requisition)
	if err == nil {
		t.Fatal("UpdateRequisitionAggregate() error = nil, want range error")
	}
	if !strings.Contains(err.Error(), "32768") || !strings.Contains(err.Error(), "32767") {
		t.Fatalf("error = %v, want SMALLINT range details", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateRequisitionAggregateInsertsWhenNoLinesWerePersisted(t *testing.T) {
	t.Parallel()

	database, mock := newRequisitionMockDB(t)
	justification, err := requisitiondomain.NewRequisitionJustification("Office setup")
	if err != nil {
		t.Fatal(err)
	}
	priority, err := requisitiondomain.NewRequisitionPriority("medium")
	if err != nil {
		t.Fatal(err)
	}

	line := updateRepositoryLine(t, "reqitem-1", 1, "itm-1", "", "Item", 1, 100)
	requisition, err := requisitiondomain.NewSubmittedRequisition(
		"req-1", "REQ-2026-10001", "usr-1", "dpt-1", priority,
		time.Now().UTC().AddDate(0, 0, 5), justification,
		[]*requisitiondomain.RequisitionItem{line},
	)
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery("SELECT id, line_number, item_id, requested_item_id").
		WillReturnRows(sqlmock.NewRows([]string{"id", "line_number", "item_id", "requested_item_id"}))
	mock.ExpectExec("UPDATE requisitions").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO requisition_items").WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewRequisitionRepository(shareddb.NewRepository(database))
	if err := repo.UpdateRequisitionAggregate(context.Background(), requisition); err != nil {
		t.Fatalf("UpdateRequisitionAggregate() unexpected error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

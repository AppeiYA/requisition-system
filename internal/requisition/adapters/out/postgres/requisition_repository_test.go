package requisitionpostgres

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"

	shareddb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/lib/pq"
)

func newRequisitionMockDB(t *testing.T) (*shareddb.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	return shareddb.New(sqlx.NewDb(sqlDB, "postgres")), mock
}

func summaryRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "reference_code", "requested_by", "department_id", "status",
		"priority", "needed_by", "justification", "total_estimated_amount",
		"created_at", "updated_at",
	})
}

func TestMapStatusHistoryWriteErrorMapsUniqueViolationToConflict(t *testing.T) {
	t.Parallel()

	err := mapStatusHistoryWriteError(&pq.Error{Code: "23505"})

	var customErr *customerrors.CustomError
	if !errors.As(err, &customErr) {
		t.Fatalf("error = %v, want a custom error", err)
	}
	if customErr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", customErr.Code, http.StatusConflict)
	}
}

func TestMapStatusHistoryWriteErrorPassesThroughOtherErrors(t *testing.T) {
	t.Parallel()

	original := errors.New("database unavailable")
	if err := mapStatusHistoryWriteError(original); !errors.Is(err, original) {
		t.Fatalf("error = %v, want %v", err, original)
	}
}

func TestListRequisitionSummariesUsesOneQueryAndMapsTotals(t *testing.T) {
	t.Parallel()

	database, mock := newRequisitionMockDB(t)
	now := time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)
	rows := summaryRows().
		AddRow("req-2", "REQ-2026-10002", "usr-2", "dpt-1", "pending",
			"medium", now.AddDate(0, 0, 5), "Office setup", int64(0), now, now).
		AddRow("req-1", "REQ-2026-10001", "usr-1", "dpt-1", "approved",
			"high", now.AddDate(0, 0, 3), "Laptops", int64(11500), now, now)

	mock.ExpectQuery(regexp.QuoteMeta(listRequisitionSummariesQuery)).
		WithArgs(100, 0).
		WillReturnRows(rows)

	repo := NewRequisitionRepository(shareddb.NewRepository(database))
	got, err := repo.ListRequisitionSummaries(context.Background(), 100, 0)
	if err != nil {
		t.Fatalf("ListRequisitionSummaries() unexpected error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("summaries = %d, want 2", len(got))
	}
	if got[0].ID != "req-2" || got[0].TotalEstimatedAmount != 0 {
		t.Fatalf("first summary = %+v, want zero total", got[0])
	}
	if got[1].ID != "req-1" || got[1].TotalEstimatedAmount != 115 {
		t.Fatalf("second summary = %+v, want 115 major units", got[1])
	}
	if got[1].ReferenceCode != "REQ-2026-10001" || got[1].RequestedBy != "usr-1" ||
		got[1].DepartmentID != "dpt-1" || got[1].Priority != "high" ||
		got[1].Status != "approved" || got[1].Justification != "Laptops" {
		t.Fatalf("header values not mapped: %+v", got[1])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListRequisitionSummariesByUserIDFiltersUserAndClamps(t *testing.T) {
	t.Parallel()

	database, mock := newRequisitionMockDB(t)
	now := time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)
	rows := summaryRows().
		AddRow("req-1", "REQ-2026-10001", "usr-123", "dpt-1", "pending",
			"medium", now.AddDate(0, 0, 5), "Office setup", int64(150000), now, now)

	mock.ExpectQuery(regexp.QuoteMeta(listRequisitionSummariesByUserIDQuery)).
		WithArgs("usr-123", 10, 20).
		WillReturnRows(rows)

	repo := NewRequisitionRepository(shareddb.NewRepository(database))
	got, err := repo.ListRequisitionSummariesByUserID(context.Background(), "usr-123", 10, 20)
	if err != nil {
		t.Fatalf("ListRequisitionSummariesByUserID() unexpected error = %v", err)
	}
	if len(got) != 1 || got[0].RequestedBy != "usr-123" {
		t.Fatalf("summaries = %+v, want one summary for usr-123", got)
	}
	if got[0].TotalEstimatedAmount != 1500 {
		t.Fatalf("total = %f, want 1500 major units", got[0].TotalEstimatedAmount)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

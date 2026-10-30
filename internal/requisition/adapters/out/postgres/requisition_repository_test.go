package requisitionpostgres

import (
	"errors"
	"net/http"
	"testing"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/lib/pq"
)

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

package requisitionhttp

import (
	"encoding/json"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

func TestCreateRequisitionRequestNeededByRequiresDate(t *testing.T) {
	t.Parallel()

	var request CreateRequisitionRequest
	if err := json.Unmarshal([]byte(`{"needed_by":"2026-11-15"}`), &request); err != nil {
		t.Fatalf("date-only needed_by should decode: %v", err)
	}
	if request.NeededBy != Date("2026-11-15") {
		t.Fatalf("NeededBy = %q, want %q", request.NeededBy, "2026-11-15")
	}
}

func TestCreateRequisitionRequestNeededByRejectsTimestamp(t *testing.T) {
	t.Parallel()

	var request CreateRequisitionRequest
	if err := json.Unmarshal([]byte(`{"needed_by":"2026-11-15T14:30:00Z"}`), &request); err == nil {
		t.Fatal("timestamp needed_by should be rejected")
	}
}

func TestCreateRequisitionRequestNeededByReportsFormatError(t *testing.T) {
	t.Parallel()

	var request CreateRequisitionRequest
	err := json.Unmarshal([]byte(`{"needed_by":"2026-28-10"}`), &request)
	if err == nil {
		t.Fatal("invalid date should be rejected")
	}

	details := response.ValidationErrors(err)
	if len(details) != 1 || details[0].Field != "needed_by" ||
		details[0].Message != "must be a date in YYYY-MM-DD format" {
		t.Fatalf("got %+v, want needed_by format detail", details)
	}
}

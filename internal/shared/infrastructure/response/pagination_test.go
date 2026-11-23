package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPaginationMeta_OmitsTotalByDefault(t *testing.T) {
	raw, err := json.Marshal(NewPaginationMeta(20, 40))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, ok := fields["total"]; ok {
		t.Fatalf("expected total to be omitted, got %s", raw)
	}
}

func TestPaginationMeta_WithTotalIncludesZero(t *testing.T) {
	raw, err := json.Marshal(NewPaginationMetaWithTotal(20, 0, 0))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	total, ok := fields["total"]
	if !ok {
		t.Fatalf("expected total to be present, got %s", raw)
	}
	if total != float64(0) {
		t.Fatalf("total = %v, want 0", total)
	}
}

func TestSuccessWithPaginationTotal_WritesTotal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	SuccessWithPaginationTotal(
		c,
		"users fetched successfully",
		http.StatusOK,
		[]any{},
		NewPaginationMetaWithTotal(20, 0, 7),
	)

	var body struct {
		Pagination PaginationMeta `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Pagination.Total == nil || *body.Pagination.Total != 7 {
		t.Fatalf("pagination.total = %v, want 7", body.Pagination.Total)
	}
	if body.Pagination.Page != 1 || body.Pagination.PageSize != 20 {
		t.Fatalf("unexpected pagination: %+v", body.Pagination)
	}
}

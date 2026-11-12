package requisitionhttp

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	"github.com/gin-gonic/gin"
)

type stubUpdateRequisition struct {
	res *domain.Requisition
	err error
	got requisitionin.UpdateRequisitionCommand
}

func (s *stubUpdateRequisition) Execute(_ context.Context, cmd requisitionin.UpdateRequisitionCommand) (*domain.Requisition, error) {
	s.got = cmd
	return s.res, s.err
}

func newRequisitionUpdateTestRouter(update requisitionin.UpdateRequisitionUsecaseInt) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	port := &requisitionin.RequisitionPortIn{UpdateRequisitionUsecase: update}
	NewRequisitionHandler(port).Register(
		router.Group("/api/v1"), requisitionAuth("usr-123"),
		func(c *gin.Context) { c.AbortWithStatus(http.StatusForbidden) },
	)
	return router
}

func Test_UpdateRequisition_MapsLineIdentity(t *testing.T) {
	t.Parallel()

	update := &stubUpdateRequisition{res: validRequisition()}
	router := newRequisitionUpdateTestRouter(update)
	body := createRequestBody(`
		{"requisition_item_id":"reqitem_1","item_id":"itm_123","quantity":2,"estimated_unit_amount":150000},
		{"requisition_item_id":"reqitem_2","requested_item":{"name":"Laptop stand","category_id":"cat_123","description":"Adjustable"},"quantity":1,"estimated_unit_amount":850000}
	`)
	rec := doReqJSON(t, router, http.MethodPatch, "/api/v1/requisitions/req_123", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if len(update.got.Lines) != 2 {
		t.Fatalf("command lines = %d, want 2", len(update.got.Lines))
	}
	if update.got.Lines[0].RequisitionItemID != "reqitem_1" || update.got.Lines[1].RequisitionItemID != "reqitem_2" {
		t.Fatalf("line identities = %+v", update.got.Lines)
	}
	if update.got.Lines[1].RequestedItem == nil || update.got.Lines[1].RequestedItem.Name != "Laptop stand" {
		t.Fatalf("requested item = %+v", update.got.Lines[1].RequestedItem)
	}
	if update.got.UserID != "usr-123" || update.got.RequisitionID != "req_123" {
		t.Fatalf("command identity = %+v", update.got)
	}
}

func Test_UpdateRequisition_AllowsLinesWithoutIdentity(t *testing.T) {
	t.Parallel()

	update := &stubUpdateRequisition{res: validRequisition()}
	router := newRequisitionUpdateTestRouter(update)
	body := createRequestBody(`{"item_id":"itm_123","quantity":1,"estimated_unit_amount":1}`)
	rec := doReqJSON(t, router, http.MethodPatch, "/api/v1/requisitions/req_123", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if update.got.Lines[0].RequisitionItemID != "" {
		t.Fatalf("line identity = %q, want empty", update.got.Lines[0].RequisitionItemID)
	}
}

func Test_UpdateRequisition_RejectsInvalidLineIdentity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines string
		field string
	}{
		{
			name:  "invalid prefix",
			lines: `{"requisition_item_id":"invalid","item_id":"itm_123","quantity":1,"estimated_unit_amount":1}`,
			field: "lines[0].requisition_item_id",
		},
		{
			name: "duplicate identity",
			lines: `{"requisition_item_id":"reqitem_1","item_id":"itm_123","quantity":1,"estimated_unit_amount":1},
				{"requisition_item_id":"reqitem_1","item_id":"itm_456","quantity":1,"estimated_unit_amount":1}`,
			field: "lines[1].requisition_item_id",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			update := &stubUpdateRequisition{res: validRequisition()}
			router := newRequisitionUpdateTestRouter(update)
			rec := doReqJSON(t, router, http.MethodPatch, "/api/v1/requisitions/req_123", createRequestBody(tt.lines))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
			}
			assertValidationField(t, rec.Body.Bytes(), tt.field)
			if len(update.got.Lines) != 0 {
				t.Fatal("usecase should not be called for invalid line identities")
			}
		})
	}
}

func Test_UpdateRequisition_DateOnlyNeededBy(t *testing.T) {
	t.Parallel()

	update := &stubUpdateRequisition{res: validRequisition()}
	router := newRequisitionUpdateTestRouter(update)
	body := `{
		"department_id":"dpt_123",
		"priority":"medium",
		"needed_by":"` + time.Now().UTC().AddDate(0, 0, 5).Format(dateOnlyLayout) + `",
		"justification":"Office setup",
		"lines":[{"requisition_item_id":"reqitem_1","item_id":"itm_123","quantity":1,"estimated_unit_amount":1}]
	}`
	rec := doReqJSON(t, router, http.MethodPatch, "/api/v1/requisitions/req_123", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if update.got.NeededBy.IsZero() {
		t.Fatal("needed_by should be mapped")
	}
}

package requisitionhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/authcontext"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/gin-gonic/gin"
)

type stubListRequisitionsByUser struct {
	res []*domain.Requisition
	err error
	got requisitionin.ListRequisitionsByUserQuery
}

func (s *stubListRequisitionsByUser) Execute(ctx context.Context, query requisitionin.ListRequisitionsByUserQuery) ([]*domain.Requisition, error) {
	s.got = query
	return s.res, s.err
}

type stubListRequisitions struct {
	res []*domain.Requisition
	err error
	got requisitionin.ListRequisitionsQuery
}

func (s *stubListRequisitions) Execute(ctx context.Context, query requisitionin.ListRequisitionsQuery) ([]*domain.Requisition, error) {
	s.got = query
	return s.res, s.err
}

type stubCreateRequisition struct {
	res *domain.Requisition
	err error
	got requisitionin.CreateRequisitionCommand
}

func (s *stubCreateRequisition) Execute(ctx context.Context, cmd requisitionin.CreateRequisitionCommand) (*domain.Requisition, error) {
	s.got = cmd
	return s.res, s.err
}

func requisitionAuth(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if userID != "" {
			c.Set(authcontext.UserIDKey, userID)
			c.Set(authcontext.UserRoleKey, "user")
		}
		c.Next()
	}
}

func requisitionAdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(authcontext.UserIDKey, "usr-admin")
		c.Set(authcontext.UserRoleKey, "admin")
		c.Next()
	}
}

func newRequisitionTestRouter(create requisitionin.CreateRequisitionUsecaseInt, listByUser requisitionin.ListRequisitionsByUserUsecaseInt, listAll requisitionin.ListRequisitionsUsecaseInt) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	port := &requisitionin.RequisitionPortIn{
		CreateRequisitionUsecase:      create,
		ListRequisitionsByUserUsecase: listByUser,
		ListRequisitionsUsecase:       listAll,
	}
	NewRequisitionHandler(port).Register(router.Group("/api/v1"), requisitionAuth("usr-123"), func(c *gin.Context) { c.AbortWithStatus(http.StatusForbidden) })
	return router
}

func newRequisitionTestRouterAdmin(create requisitionin.CreateRequisitionUsecaseInt, listByUser requisitionin.ListRequisitionsByUserUsecaseInt, listAll requisitionin.ListRequisitionsUsecaseInt) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	port := &requisitionin.RequisitionPortIn{
		CreateRequisitionUsecase:      create,
		ListRequisitionsByUserUsecase: listByUser,
		ListRequisitionsUsecase:       listAll,
	}
	NewRequisitionHandler(port).Register(router.Group("/api/v1"), requisitionAdminAuth(), func(c *gin.Context) { c.Next() })
	return router
}

func doReqJSON(t *testing.T, router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func validRequisition() *domain.Requisition {
	catalogueName, _ := domain.NewItemName("Pens")
	requestedName, _ := domain.NewItemName("Laptop stand")
	requestedDescription, _ := domain.NewItemDescription("Adjustable laptop stand")
	quantity, _ := domain.NewRequisitionQuantity(2)
	requestedQuantity, _ := domain.NewRequisitionQuantity(1)
	catalogueAmount, _ := shareddomain.NewNairaMoneyFromMinorUnits(150000)
	requestedAmount, _ := shareddomain.NewNairaMoneyFromMinorUnits(850000)
	justification, _ := domain.NewRequisitionJustification("Needed for office")
	priority, _ := domain.NewRequisitionPriority("medium")

	firstLine, _ := domain.NewRequisitionItem(
		"reqitem_1", "req_123", 1, "itm_123", "", catalogueName,
		quantity, catalogueAmount, "Optional note",
	)
	secondLine, _ := domain.NewRequisitionItem(
		"reqitem_2", "req_123", 2, "", "ritem_123", requestedName,
		requestedQuantity, requestedAmount, "",
	)
	requestedItem, _ := domain.NewRequestedItem(
		"ritem_123", requestedName, "cat_123", requestedDescription, "usr_123",
	)

	requisition, err := domain.NewSubmittedRequisitionWithRequestedItems(
		"req_123", "REQ-2026-10001", "usr_123", "dpt_123",
		priority, time.Now().UTC().AddDate(0, 0, 5), justification,
		[]*domain.RequisitionItem{firstLine, secondLine}, []*domain.RequestedItem{requestedItem},
	)
	if err != nil {
		panic(err)
	}
	history, err := requisition.InitialStatusHistory(
		"rsh_1", "usr_123", domain.ActorRoleStaff, time.Now().UTC(),
	)
	if err != nil {
		panic(err)
	}
	if err := requisition.RecordInitialStatusHistory(history); err != nil {
		panic(err)
	}
	return requisition
}

func createRequestBody(lines string) string {
	return `{
		"department_id": "dpt_123",
		"priority": "medium",
		"needed_by": "` + time.Now().UTC().AddDate(0, 0, 5).Format(dateOnlyLayout) + `",
		"justification": "Office setup",
		"lines": [` + lines + `]
	}`
}

func decodeCreatedRequisition(t *testing.T, body []byte) RequisitionResponse {
	t.Helper()
	var response struct {
		Data RequisitionResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return response.Data
}

func Test_CreateRequisition_MultipleCatalogueLines(t *testing.T) {
	t.Parallel()
	create := &stubCreateRequisition{res: validRequisition()}
	router := newRequisitionTestRouter(create, &stubListRequisitionsByUser{}, &stubListRequisitions{})

	body := createRequestBody(`
		{"item_id":"itm_123","quantity":2,"estimated_unit_amount":150000,"notes":"Optional note"},
		{"item_id":"itm_456","quantity":3,"estimated_unit_amount":50000}
	`)
	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", rec.Code, rec.Body.String())
	}
	if len(create.got.Lines) != 2 {
		t.Fatalf("command lines = %d, want 2", len(create.got.Lines))
	}
	if create.got.DepartmentID != "dpt_123" || create.got.UserID != "usr-123" {
		t.Fatalf("command header = %+v", create.got)
	}
	if create.got.Lines[0].ItemID != "itm_123" || create.got.Lines[0].EstimatedUnitAmount != 150000 {
		t.Fatalf("first command line = %+v", create.got.Lines[0])
	}
	if create.got.Lines[1].ItemID != "itm_456" || create.got.Lines[1].Quantity != 3 {
		t.Fatalf("second command line = %+v", create.got.Lines[1])
	}
}

func Test_CreateRequisition_CatalogueAndRequestedLines(t *testing.T) {
	t.Parallel()
	create := &stubCreateRequisition{res: validRequisition()}
	router := newRequisitionTestRouter(create, &stubListRequisitionsByUser{}, &stubListRequisitions{})

	body := createRequestBody(`
		{"item_id":"itm_123","quantity":2,"estimated_unit_amount":150000,"notes":"Optional note"},
		{"requested_item":{"name":"Laptop stand","category_id":"cat_123","description":"Adjustable laptop stand"},"quantity":1,"estimated_unit_amount":850000}
	`)
	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", rec.Code, rec.Body.String())
	}
	if len(create.got.Lines) != 2 || create.got.Lines[1].RequestedItem == nil {
		t.Fatalf("command lines = %+v", create.got.Lines)
	}
	requested := create.got.Lines[1].RequestedItem
	if requested.Name != "Laptop stand" || requested.CategoryID != "cat_123" || requested.Description != "Adjustable laptop stand" {
		t.Fatalf("requested item = %+v", requested)
	}
}

func Test_CreateRequisition_AcceptsCamelCaseAliases(t *testing.T) {
	t.Parallel()
	create := &stubCreateRequisition{res: validRequisition()}
	router := newRequisitionTestRouter(create, &stubListRequisitionsByUser{}, &stubListRequisitions{})

	body := `{
		"departmentId": "dpt_123",
		"priority": "medium",
		"neededBy": "` + time.Now().UTC().AddDate(0, 0, 5).Format(dateOnlyLayout) + `",
		"justification": "Office setup",
		"lines": [
			{"itemId":"itm_123","quantity":2,"estimatedUnitAmount":150000},
			{"requestedItem":{"name":"Laptop stand","categoryId":"cat_123","description":"Adjustable laptop stand"},"quantity":1,"estimatedUnitAmount":850000}
		]
	}`
	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", rec.Code, rec.Body.String())
	}
	if create.got.DepartmentID != "dpt_123" || create.got.Lines[0].ItemID != "itm_123" {
		t.Fatalf("command = %+v", create.got)
	}
	if create.got.Lines[1].RequestedItem == nil || create.got.Lines[1].RequestedItem.CategoryID != "cat_123" {
		t.Fatalf("requested item = %+v", create.got.Lines[1].RequestedItem)
	}
}

func Test_CreateRequisition_RejectsEmptyLines(t *testing.T) {
	t.Parallel()
	router := newRequisitionTestRouter(&stubCreateRequisition{}, &stubListRequisitionsByUser{}, &stubListRequisitions{})
	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions", createRequestBody(""))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
	assertValidationField(t, rec.Body.Bytes(), "lines")
}

func Test_CreateRequisition_RejectsInvalidLineExclusivity(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		line string
	}{
		{
			name: "both",
			line: `{"item_id":"itm_123","requested_item":{"name":"Laptop stand","category_id":"cat_123","description":"Adjustable"},"quantity":1,"estimated_unit_amount":100}`,
		},
		{
			name: "neither",
			line: `{"quantity":1,"estimated_unit_amount":100}`,
		},
	}
	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			create := &stubCreateRequisition{res: validRequisition()}
			router := newRequisitionTestRouter(create, &stubListRequisitionsByUser{}, &stubListRequisitions{})
			rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions", createRequestBody(tt.line))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
			}
			assertValidationField(t, rec.Body.Bytes(), "lines[0]")
			if len(create.got.Lines) != 0 {
				t.Fatal("usecase should not be called for invalid lines")
			}
		})
	}
}

func Test_CreateRequisition_RejectsDuplicateCatalogueItems(t *testing.T) {
	t.Parallel()
	router := newRequisitionTestRouter(&stubCreateRequisition{res: validRequisition()}, &stubListRequisitionsByUser{}, &stubListRequisitions{})
	body := createRequestBody(`
		{"item_id":"itm_123","quantity":1,"estimated_unit_amount":100},
		{"item_id":"itm_123","quantity":2,"estimated_unit_amount":100}
	`)
	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
	assertValidationField(t, rec.Body.Bytes(), "lines[1].item_id")
}

func Test_CreateRequisition_RejectsInvalidQuantityAndAmount(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		line  string
		field string
	}{
		{
			name:  "zero quantity",
			line:  `{"item_id":"itm_123","quantity":0,"estimated_unit_amount":100}`,
			field: "lines[0].quantity",
		},
		{
			name:  "negative quantity",
			line:  `{"item_id":"itm_123","quantity":-1,"estimated_unit_amount":100}`,
			field: "lines[0].quantity",
		},
		{
			name:  "negative amount",
			line:  `{"item_id":"itm_123","quantity":1,"estimated_unit_amount":-1}`,
			field: "lines[0].estimated_unit_amount",
		},
	}
	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			router := newRequisitionTestRouter(&stubCreateRequisition{res: validRequisition()}, &stubListRequisitionsByUser{}, &stubListRequisitions{})
			rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions", createRequestBody(tt.line))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
			}
			assertValidationField(t, rec.Body.Bytes(), tt.field)
		})
	}
}

func Test_CreateRequisition_RejectsInvalidIdentifiersAndNeededBy(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		body  string
		field string
	}{
		{
			name:  "invalid department",
			body:  `{"department_id":"invalid","priority":"medium","needed_by":"` + time.Now().UTC().AddDate(0, 0, 5).Format(dateOnlyLayout) + `","justification":"x","lines":[{"item_id":"itm_123","quantity":1,"estimated_unit_amount":1}]}`,
			field: "department_id",
		},
		{
			name:  "invalid item",
			body:  `{"department_id":"dpt_123","priority":"medium","needed_by":"` + time.Now().UTC().AddDate(0, 0, 5).Format(dateOnlyLayout) + `","justification":"x","lines":[{"item_id":"invalid","quantity":1,"estimated_unit_amount":1}]}`,
			field: "lines[0].item_id",
		},
		{
			name:  "invalid category",
			body:  `{"department_id":"dpt_123","priority":"medium","needed_by":"` + time.Now().UTC().AddDate(0, 0, 5).Format(dateOnlyLayout) + `","justification":"x","lines":[{"requested_item":{"name":"Stand","category_id":"invalid","description":"Adjustable"},"quantity":1,"estimated_unit_amount":1}]}`,
			field: "lines[0].requested_item.category_id",
		},
		{
			name:  "invalid priority",
			body:  `{"department_id":"dpt_123","priority":"extreme","needed_by":"` + time.Now().UTC().AddDate(0, 0, 5).Format(dateOnlyLayout) + `","justification":"x","lines":[{"item_id":"itm_123","quantity":1,"estimated_unit_amount":1}]}`,
			field: "priority",
		},
		{
			name:  "past needed by",
			body:  `{"department_id":"dpt_123","priority":"medium","needed_by":"2000-01-01","justification":"x","lines":[{"item_id":"itm_123","quantity":1,"estimated_unit_amount":1}]}`,
			field: "needed_by",
		},
	}
	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			router := newRequisitionTestRouter(&stubCreateRequisition{res: validRequisition()}, &stubListRequisitionsByUser{}, &stubListRequisitions{})
			rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions", tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
			}
			assertValidationField(t, rec.Body.Bytes(), tt.field)
		})
	}
}

func Test_CreateRequisition_MapsMinorUnitAmounts(t *testing.T) {
	t.Parallel()
	create := &stubCreateRequisition{res: validRequisition()}
	router := newRequisitionTestRouter(create, &stubListRequisitionsByUser{}, &stubListRequisitions{})
	body := createRequestBody(`{"item_id":"itm_123","quantity":2,"estimated_unit_amount":150000}`)
	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", rec.Code, rec.Body.String())
	}
	if got := create.got.Lines[0].EstimatedUnitAmount; got != 150000 {
		t.Fatalf("command estimated unit amount = %f, want 150000 major units", got)
	}
}

func Test_CreateRequisition_SerializesLinesTotalsAndStatusHistory(t *testing.T) {
	t.Parallel()
	create := &stubCreateRequisition{res: validRequisition()}
	router := newRequisitionTestRouter(create, &stubListRequisitionsByUser{}, &stubListRequisitions{})
	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions", createRequestBody(`{"item_id":"itm_123","quantity":2,"estimated_unit_amount":150000}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", rec.Code, rec.Body.String())
	}

	var raw struct {
		Data struct {
			ReferenceCode        string  `json:"reference_code"`
			Status               string  `json:"status"`
			TotalEstimatedAmount float64 `json:"total_estimated_amount"`
			Lines                []struct {
				ID                  string  `json:"id"`
				LineNumber          int     `json:"line_number"`
				ItemID              string  `json:"item_id"`
				RequestedItemID     string  `json:"requested_item_id"`
				ItemName            string  `json:"item_name"`
				Quantity            int64   `json:"quantity"`
				EstimatedUnitAmount float64 `json:"estimated_unit_amount"`
				Notes               string  `json:"notes"`
				LineTotal           float64 `json:"line_total"`
			} `json:"lines"`
			StatusHistory []struct {
				Sequence   int     `json:"sequence"`
				FromStatus *string `json:"from_status"`
				ToStatus   string  `json:"to_status"`
				ActedBy    string  `json:"acted_by"`
				ActorRole  string  `json:"actor_role"`
				Reason     string  `json:"reason"`
				CreatedAt  string  `json:"created_at"`
			} `json:"status_history"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if raw.Data.ReferenceCode != "REQ-2026-10001" || raw.Data.Status != "pending" {
		t.Fatalf("header = %+v", raw.Data)
	}
	if raw.Data.TotalEstimatedAmount != 11500 {
		t.Fatalf("total = %f, want 11500", raw.Data.TotalEstimatedAmount)
	}
	if len(raw.Data.Lines) != 2 {
		t.Fatalf("lines = %d, want 2", len(raw.Data.Lines))
	}
	if raw.Data.Lines[0].LineTotal != 3000 || raw.Data.Lines[1].LineTotal != 8500 {
		t.Fatalf("line totals = %f, %f", raw.Data.Lines[0].LineTotal, raw.Data.Lines[1].LineTotal)
	}
	if raw.Data.Lines[0].ItemID != "itm_123" || raw.Data.Lines[1].RequestedItemID != "ritem_123" {
		t.Fatalf("line sources = %+v", raw.Data.Lines)
	}
	if raw.Data.Lines[1].ItemName != "Laptop stand" || raw.Data.Lines[1].Quantity != 1 {
		t.Fatalf("requested line = %+v", raw.Data.Lines[1])
	}
	if len(raw.Data.StatusHistory) != 1 {
		t.Fatalf("status history = %d, want 1", len(raw.Data.StatusHistory))
	}
	history := raw.Data.StatusHistory[0]
	if history.Sequence != 1 || history.FromStatus != nil || history.ToStatus != "pending" ||
		history.ActedBy != "usr_123" || history.ActorRole != "staff" || history.CreatedAt == "" {
		t.Fatalf("status history = %+v", history)
	}
}

func Test_CreateRequisition_PropagatesUsecaseError(t *testing.T) {
	t.Parallel()
	create := &stubCreateRequisition{err: customerrors.NewBadRequestError("selected item is inactive")}
	router := newRequisitionTestRouter(create, &stubListRequisitionsByUser{}, &stubListRequisitions{})
	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions", createRequestBody(`{"item_id":"itm_123","quantity":1,"estimated_unit_amount":1}`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Success || body.Message != "selected item is inactive" {
		t.Fatalf("error body = %+v", body)
	}
}

func Test_CreateRequisition_Unauthenticated(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	port := &requisitionin.RequisitionPortIn{
		CreateRequisitionUsecase:      &stubCreateRequisition{},
		ListRequisitionsByUserUsecase: &stubListRequisitionsByUser{},
		ListRequisitionsUsecase:       &stubListRequisitions{},
	}
	NewRequisitionHandler(port).Register(router.Group("/api/v1"), func(c *gin.Context) { c.Next() }, func(c *gin.Context) { c.Next() })
	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions", createRequestBody(`{"item_id":"itm_123","quantity":1,"estimated_unit_amount":1}`))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401: %s", rec.Code, rec.Body.String())
	}
}

func assertValidationField(t *testing.T, body []byte, field string) {
	t.Helper()
	var response struct {
		Error   string `json:"error"`
		Details []struct {
			Field   string `json:"field"`
			Message string `json:"message"`
		} `json:"details"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("unmarshal validation response: %v", err)
	}
	if response.Error != "validation_error" {
		t.Fatalf("error = %q, want validation_error", response.Error)
	}
	for _, detail := range response.Details {
		if detail.Field == field || strings.HasPrefix(detail.Field, field+".") {
			return
		}
	}
	t.Fatalf("validation details = %+v, want field %q", response.Details, field)
}

func Test_ListMyRequisitions_Success(t *testing.T) {
	t.Parallel()
	listByUser := &stubListRequisitionsByUser{res: []*domain.Requisition{validRequisition()}}
	router := newRequisitionTestRouter(&stubCreateRequisition{}, listByUser, &stubListRequisitions{})
	rec := doReqJSON(t, router, http.MethodGet, "/api/v1/requisitions", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if listByUser.got.UserID != "usr-123" {
		t.Errorf("expected user id usr-123, got %s", listByUser.got.UserID)
	}
	if listByUser.got.Limit != 100 || listByUser.got.Offset != 0 {
		t.Errorf("expected default pagination, got limit=%d offset=%d", listByUser.got.Limit, listByUser.got.Offset)
	}
	var resp struct {
		Data []*RequisitionResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 requisition, got %d", len(resp.Data))
	}
	if resp.Data[0] == nil {
		t.Fatal("expected non-nil requisition")
	}
	if len(resp.Data[0].Lines) != 2 || resp.Data[0].TotalEstimatedAmount != 11500 {
		t.Fatalf("unexpected aggregate response: %+v", resp.Data[0])
	}
}

func Test_ListMyRequisitions_Empty(t *testing.T) {
	t.Parallel()
	listByUser := &stubListRequisitionsByUser{res: []*domain.Requisition{}}
	router := newRequisitionTestRouter(&stubCreateRequisition{}, listByUser, &stubListRequisitions{})
	rec := doReqJSON(t, router, http.MethodGet, "/api/v1/requisitions", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if string(resp["data"]) != "[]" {
		t.Fatalf("expected data: [], got %s", string(resp["data"]))
	}
}

func Test_ListMyRequisitions_Pagination(t *testing.T) {
	t.Parallel()
	listByUser := &stubListRequisitionsByUser{res: []*domain.Requisition{}}
	router := newRequisitionTestRouter(&stubCreateRequisition{}, listByUser, &stubListRequisitions{})
	rec := doReqJSON(t, router, http.MethodGet, "/api/v1/requisitions?limit=10&offset=20", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if listByUser.got.Limit != 10 || listByUser.got.Offset != 20 {
		t.Errorf("expected limit=10 offset=20, got limit=%d offset=%d", listByUser.got.Limit, listByUser.got.Offset)
	}
}

func Test_ListAllRequisitions_Admin(t *testing.T) {
	t.Parallel()
	listAll := &stubListRequisitions{res: []*domain.Requisition{validRequisition()}}
	router := newRequisitionTestRouterAdmin(&stubCreateRequisition{}, &stubListRequisitionsByUser{}, listAll)
	rec := doReqJSON(t, router, http.MethodGet, "/api/v1/requisitions/all", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if listAll.got.Limit != 100 || listAll.got.Offset != 0 {
		t.Errorf("expected default pagination, got limit=%d offset=%d", listAll.got.Limit, listAll.got.Offset)
	}
}

func Test_ListAllRequisitions_NonAdminForbidden(t *testing.T) {
	t.Parallel()
	router := newRequisitionTestRouter(&stubCreateRequisition{}, &stubListRequisitionsByUser{}, &stubListRequisitions{})
	rec := doReqJSON(t, router, http.MethodGet, "/api/v1/requisitions/all", "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", rec.Code, rec.Body.String())
	}
}

func Test_ListMyRequisitions_Unauthenticated(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	port := &requisitionin.RequisitionPortIn{
		CreateRequisitionUsecase:      &stubCreateRequisition{},
		ListRequisitionsByUserUsecase: &stubListRequisitionsByUser{},
		ListRequisitionsUsecase:       &stubListRequisitions{},
	}
	// no user in context
	NewRequisitionHandler(port).Register(router.Group("/api/v1"), func(c *gin.Context) { c.Next() }, func(c *gin.Context) { c.Next() })
	rec := doReqJSON(t, router, http.MethodGet, "/api/v1/requisitions", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401: %s", rec.Code, rec.Body.String())
	}
}

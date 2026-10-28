package requisitionhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/authcontext"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
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
	name, _ := domain.NewItemName("Pens")
	desc, _ := domain.NewItemDescription("Blue pens")
	qty, _ := domain.NewRequisitionQuantity(10)
	cost, _ := shareddomain.NewNairaMoney(1000)
	just, _ := domain.NewRequisitionJustification("Needed for office")
	priority, _ := domain.NewRequisitionPriority("normal")
	r, err := domain.NewRequisition("req-1234", "usr-123", "dpt-1234", "", "cat-1234", name, desc, qty, cost, just, priority, time.Date(2099, 12, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	return r
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

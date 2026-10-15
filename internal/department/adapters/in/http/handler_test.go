package departmenthttp

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/gin-gonic/gin"
)

type stubCreateDepartment struct {
	res *departmentin.CreateDepartmentResult
	err error
	got departmentin.CreateDepartmentCommand
}

func (s *stubCreateDepartment) Execute(ctx context.Context, cmd departmentin.CreateDepartmentCommand) (*departmentin.CreateDepartmentResult, error) {
	s.got = cmd
	return s.res, s.err
}

type stubGetDepartmentByID struct {
	res   *departmentin.DepartmentResult
	err   error
	gotID string
}

func (s *stubGetDepartmentByID) Execute(ctx context.Context, id string) (*departmentin.DepartmentResult, error) {
	s.gotID = id
	return s.res, s.err
}

func withNoAuth() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

func newTestRouter(create departmentin.CreateDepartmentUsecaseInt, get departmentin.GetDepartmentByIDUsecaseInt) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	NewDepartmentHandler(create, get).Register(router.Group("/api/v1"), withNoAuth())

	return router
}

func doJSON(t *testing.T, router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func Test_CreateDepartment_Success(t *testing.T) {
	t.Parallel()

	create := &stubCreateDepartment{res: &departmentin.CreateDepartmentResult{ID: "dpt-1", Name: "Finance", Code: "FIN", Status: "active"}}
	router := newTestRouter(create, &stubGetDepartmentByID{})

	rec := doJSON(t, router, http.MethodPost, "/api/v1/admin/departments", `{"name":"Finance","description":"Handles payments","code":"FIN"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", rec.Code, rec.Body.String())
	}
	if create.got.Name != "Finance" || create.got.Description != "Handles payments" || create.got.Code != "FIN" {
		t.Errorf("expected fields to be forwarded, got %+v", create.got)
	}
}

func Test_CreateDepartment_InvalidBody(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&stubCreateDepartment{}, &stubGetDepartmentByID{})

	rec := doJSON(t, router, http.MethodPost, "/api/v1/admin/departments", `{"name":"Finance"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

func Test_CreateDepartment_Conflict(t *testing.T) {
	t.Parallel()

	create := &stubCreateDepartment{err: customerrors.NewConflictError("department with this code already exists")}
	router := newTestRouter(create, &stubGetDepartmentByID{})

	rec := doJSON(t, router, http.MethodPost, "/api/v1/admin/departments", `{"name":"Finance","description":"Handles payments","code":"FIN"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", rec.Code, rec.Body.String())
	}
}

func Test_GetDepartment_Success(t *testing.T) {
	t.Parallel()

	get := &stubGetDepartmentByID{res: &departmentin.DepartmentResult{ID: "dpt-1", Name: "Finance", Status: "active"}}
	router := newTestRouter(&stubCreateDepartment{}, get)

	rec := doJSON(t, router, http.MethodGet, "/api/v1/admin/departments/dpt-1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if get.gotID != "dpt-1" {
		t.Errorf("expected id %q, got %q", "dpt-1", get.gotID)
	}
}

func Test_GetDepartment_NotFound(t *testing.T) {
	t.Parallel()

	get := &stubGetDepartmentByID{err: customerrors.NewNotFoundError("department not found")}
	router := newTestRouter(&stubCreateDepartment{}, get)

	rec := doJSON(t, router, http.MethodGet, "/api/v1/admin/departments/dpt-missing", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

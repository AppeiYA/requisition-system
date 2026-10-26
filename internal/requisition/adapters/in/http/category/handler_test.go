package category

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionmocks "github.com/AppeiYA/requisition-system/internal/requisition/mocks"
	categoryin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/category"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/gin-gonic/gin"
)

type stubCreateItemCategory struct {
	res *domain.ItemCategory
	err error
	got categoryin.CreateItemCategoryCommand
}

func (s *stubCreateItemCategory) Execute(ctx context.Context, cmd categoryin.CreateItemCategoryCommand) (*domain.ItemCategory, error) {
	s.got = cmd
	return s.res, s.err
}

type stubGetItemCategory struct {
	res   *domain.ItemCategory
	err   error
	gotID string
}

func (s *stubGetItemCategory) Execute(ctx context.Context, id string) (*domain.ItemCategory, error) {
	s.gotID = id
	return s.res, s.err
}

type stubListItemCategories struct {
	res []*domain.ItemCategory
	err error
	got categoryin.ListItemCategoriesQuery
}

func (s *stubListItemCategories) Execute(ctx context.Context, query categoryin.ListItemCategoriesQuery) ([]*domain.ItemCategory, error) {
	s.got = query
	return s.res, s.err
}

type stubUpdateItemCategory struct {
	res *domain.ItemCategory
	err error
	got categoryin.UpdateItemCategoryCommand
}

func (s *stubUpdateItemCategory) Execute(ctx context.Context, cmd categoryin.UpdateItemCategoryCommand) (*domain.ItemCategory, error) {
	s.got = cmd
	return s.res, s.err
}

type stubActivateItemCategory struct {
	res   *domain.ItemCategory
	err   error
	gotID string
}

func (s *stubActivateItemCategory) Execute(ctx context.Context, id string) (*domain.ItemCategory, error) {
	s.gotID = id
	return s.res, s.err
}

type stubDeactivateItemCategory struct {
	res   *domain.ItemCategory
	err   error
	gotID string
}

func (s *stubDeactivateItemCategory) Execute(ctx context.Context, id string) (*domain.ItemCategory, error) {
	s.gotID = id
	return s.res, s.err
}

func allow() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

func deny(status int) gin.HandlerFunc {
	return func(c *gin.Context) { c.AbortWithStatus(status) }
}

func newTestRouter(
	create categoryin.CreateItemCategoryUsecaseInt,
	get categoryin.GetItemCategoryUsecaseInt,
	list categoryin.ListItemCategoriesUsecaseInt,
	update categoryin.UpdateItemCategoryUsecaseInt,
	authMiddleware gin.HandlerFunc,
	adminOnlyMiddleware gin.HandlerFunc,
) *gin.Engine {
	return newTestRouterWithStatus(
		create,
		get,
		list,
		update,
		&stubActivateItemCategory{},
		&stubDeactivateItemCategory{},
		authMiddleware,
		adminOnlyMiddleware,
	)
}

func newTestRouterWithStatus(
	create categoryin.CreateItemCategoryUsecaseInt,
	get categoryin.GetItemCategoryUsecaseInt,
	list categoryin.ListItemCategoriesUsecaseInt,
	update categoryin.UpdateItemCategoryUsecaseInt,
	activate categoryin.ActivateItemCategoryUsecaseInt,
	deactivate categoryin.DeactivateItemCategoryUsecaseInt,
	authMiddleware gin.HandlerFunc,
	adminOnlyMiddleware gin.HandlerFunc,
) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	port := &categoryin.ItemCategoryPortIn{
		CreateItemCategoryUsecase:     create,
		GetItemCategoryUsecase:        get,
		ListItemCategoriesUsecase:     list,
		UpdateItemCategoryUsecase:     update,
		ActivateItemCategoryUsecase:   activate,
		DeactivateItemCategoryUsecase: deactivate,
	}
	NewItemCategoryHandler(port).Register(router.Group("/api/v1"), authMiddleware, adminOnlyMiddleware)

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

func Test_CreateItemCategory_Success(t *testing.T) {
	t.Parallel()

	create := &stubCreateItemCategory{res: requisitionmocks.ValidItemCategory()}
	router := newTestRouter(create, &stubGetItemCategory{}, &stubListItemCategories{}, &stubUpdateItemCategory{}, allow(), allow())

	rec := doJSON(t, router, http.MethodPost, "/api/v1/categories", `{"name":"Office Supplies","description":"General office supplies"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", rec.Code, rec.Body.String())
	}
	if create.got.Name != "Office Supplies" || create.got.Description != "General office supplies" {
		t.Errorf("expected fields to be forwarded, got %+v", create.got)
	}
}

func Test_CreateItemCategory_InvalidBody(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&stubCreateItemCategory{}, &stubGetItemCategory{}, &stubListItemCategories{}, &stubUpdateItemCategory{}, allow(), allow())

	rec := doJSON(t, router, http.MethodPost, "/api/v1/categories", `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

func Test_CreateItemCategory_Conflict(t *testing.T) {
	t.Parallel()

	create := &stubCreateItemCategory{err: customerrors.NewConflictError("item category with this name already exists")}
	router := newTestRouter(create, &stubGetItemCategory{}, &stubListItemCategories{}, &stubUpdateItemCategory{}, allow(), allow())

	rec := doJSON(t, router, http.MethodPost, "/api/v1/categories", `{"name":"Office Supplies"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", rec.Code, rec.Body.String())
	}
}

func Test_CreateItemCategory_RequiresAdmin(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&stubCreateItemCategory{}, &stubGetItemCategory{}, &stubListItemCategories{}, &stubUpdateItemCategory{}, allow(), deny(http.StatusForbidden))

	rec := doJSON(t, router, http.MethodPost, "/api/v1/categories", `{"name":"Office Supplies"}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", rec.Code, rec.Body.String())
	}
}

func Test_ListItemCategories_Success(t *testing.T) {
	t.Parallel()

	list := &stubListItemCategories{res: []*domain.ItemCategory{requisitionmocks.ValidItemCategory()}}
	router := newTestRouter(&stubCreateItemCategory{}, &stubGetItemCategory{}, list, &stubUpdateItemCategory{}, allow(), allow())

	rec := doJSON(t, router, http.MethodGet, "/api/v1/categories?limit=10&offset=2", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if list.got.Limit != 10 || list.got.Offset != 2 {
		t.Errorf("expected limit=10 offset=2, got %+v", list.got)
	}

	var body struct {
		Success bool `json:"success"`
		Data    []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success || len(body.Data) != 1 {
		t.Fatalf("unexpected response: %s", rec.Body.String())
	}
	if body.Data[0].ID != requisitionmocks.DefaultItemCategoryID {
		t.Errorf("expected id %q, got %q", requisitionmocks.DefaultItemCategoryID, body.Data[0].ID)
	}
}

func Test_ListItemCategories_InvalidLimit(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&stubCreateItemCategory{}, &stubGetItemCategory{}, &stubListItemCategories{}, &stubUpdateItemCategory{}, allow(), allow())

	rec := doJSON(t, router, http.MethodGet, "/api/v1/categories?limit=101", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

func Test_ListItemCategories_RequiresAuth(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&stubCreateItemCategory{}, &stubGetItemCategory{}, &stubListItemCategories{}, &stubUpdateItemCategory{}, deny(http.StatusUnauthorized), allow())

	rec := doJSON(t, router, http.MethodGet, "/api/v1/categories", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401: %s", rec.Code, rec.Body.String())
	}
}

func Test_GetItemCategory_Success(t *testing.T) {
	t.Parallel()

	get := &stubGetItemCategory{res: requisitionmocks.ValidItemCategory()}
	router := newTestRouter(&stubCreateItemCategory{}, get, &stubListItemCategories{}, &stubUpdateItemCategory{}, allow(), allow())

	rec := doJSON(t, router, http.MethodGet, "/api/v1/categories/cat-1234", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if get.gotID != requisitionmocks.DefaultItemCategoryID {
		t.Errorf("expected id %q, got %q", requisitionmocks.DefaultItemCategoryID, get.gotID)
	}
}

func Test_GetItemCategory_NotFound(t *testing.T) {
	t.Parallel()

	get := &stubGetItemCategory{err: customerrors.NewNotFoundError("item category not found")}
	router := newTestRouter(&stubCreateItemCategory{}, get, &stubListItemCategories{}, &stubUpdateItemCategory{}, allow(), allow())

	rec := doJSON(t, router, http.MethodGet, "/api/v1/categories/cat-missing", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func Test_UpdateItemCategory_SuccessWithOptionalFields(t *testing.T) {
	t.Parallel()

	update := &stubUpdateItemCategory{res: requisitionmocks.ValidItemCategory()}
	router := newTestRouter(&stubCreateItemCategory{}, &stubGetItemCategory{}, &stubListItemCategories{}, update, allow(), allow())

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/categories/cat-1234", `{"name":"Hardware"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if update.got.ID != requisitionmocks.DefaultItemCategoryID {
		t.Errorf("expected id %q, got %q", requisitionmocks.DefaultItemCategoryID, update.got.ID)
	}
	if update.got.Name == nil || *update.got.Name != "Hardware" {
		t.Errorf("expected name to be forwarded, got %+v", update.got.Name)
	}
	if update.got.Description != nil {
		t.Errorf("expected description to be nil, got %q", *update.got.Description)
	}
}

func Test_UpdateItemCategory_EmptyBody(t *testing.T) {
	t.Parallel()

	update := &stubUpdateItemCategory{err: domain.ErrNoCategoryUpdateFields}
	router := newTestRouter(&stubCreateItemCategory{}, &stubGetItemCategory{}, &stubListItemCategories{}, update, allow(), allow())

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/categories/cat-1234", `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
	if update.got.Name != nil || update.got.Description != nil {
		t.Errorf("expected all update fields to be nil, got %+v", update.got)
	}
}

func Test_UpdateItemCategory_RequiresAdmin(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&stubCreateItemCategory{}, &stubGetItemCategory{}, &stubListItemCategories{}, &stubUpdateItemCategory{}, allow(), deny(http.StatusForbidden))

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/categories/cat-1234", `{"name":"Hardware"}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", rec.Code, rec.Body.String())
	}
}

func Test_ActivateItemCategory_Success(t *testing.T) {
	t.Parallel()

	activate := &stubActivateItemCategory{res: requisitionmocks.ValidItemCategory()}
	router := newTestRouterWithStatus(
		&stubCreateItemCategory{},
		&stubGetItemCategory{},
		&stubListItemCategories{},
		&stubUpdateItemCategory{},
		activate,
		&stubDeactivateItemCategory{},
		allow(),
		allow(),
	)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/categories/cat-1234/activate", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if activate.gotID != requisitionmocks.DefaultItemCategoryID {
		t.Errorf("expected id %q, got %q", requisitionmocks.DefaultItemCategoryID, activate.gotID)
	}
}

func Test_DeactivateItemCategory_Success(t *testing.T) {
	t.Parallel()

	deactivate := &stubDeactivateItemCategory{res: requisitionmocks.ValidItemCategory()}
	router := newTestRouterWithStatus(
		&stubCreateItemCategory{},
		&stubGetItemCategory{},
		&stubListItemCategories{},
		&stubUpdateItemCategory{},
		&stubActivateItemCategory{},
		deactivate,
		allow(),
		allow(),
	)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/categories/cat-1234/deactivate", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if deactivate.gotID != requisitionmocks.DefaultItemCategoryID {
		t.Errorf("expected id %q, got %q", requisitionmocks.DefaultItemCategoryID, deactivate.gotID)
	}
}

func Test_ActivateDeactivateItemCategory_RequireAdmin(t *testing.T) {
	t.Parallel()

	for _, path := range []string{
		"/api/v1/categories/cat-1234/activate",
		"/api/v1/categories/cat-1234/deactivate",
	} {
		router := newTestRouterWithStatus(
			&stubCreateItemCategory{},
			&stubGetItemCategory{},
			&stubListItemCategories{},
			&stubUpdateItemCategory{},
			&stubActivateItemCategory{},
			&stubDeactivateItemCategory{},
			allow(),
			deny(http.StatusForbidden),
		)

		rec := doJSON(t, router, http.MethodPatch, path, "")
		if rec.Code != http.StatusForbidden {
			t.Fatalf("%s status = %d, want 403: %s", path, rec.Code, rec.Body.String())
		}
	}
}

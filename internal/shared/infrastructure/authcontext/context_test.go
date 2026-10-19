package authcontext

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthContext_Gin(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	SetUser(c, "usr-123", "admin")

	id, ok := GetUserID(c)
	if !ok || id != "usr-123" {
		t.Fatalf("expected id usr-123, got %q (ok=%v)", id, ok)
	}

	role, ok := GetUserRole(c)
	if !ok || role != "admin" {
		t.Fatalf("expected role admin, got %q (ok=%v)", role, ok)
	}

	info, ok := FromContext(c.Request.Context())
	if !ok || info.ID != "usr-123" || info.Role != "admin" {
		t.Fatalf("expected FromContext usr-123/admin, got %+v (ok=%v)", info, ok)
	}
}

func TestAuthContext_StdContext(t *testing.T) {
	t.Parallel()

	ctx := WithUser(context.Background(), "usr-456", "user")
	info, ok := FromContext(ctx)
	if !ok || info.ID != "usr-456" || info.Role != "user" {
		t.Fatalf("expected usr-456/user, got %+v (ok=%v)", info, ok)
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

	id, ok := GetUserID(c)
	if !ok || id != "usr-456" {
		t.Fatalf("expected fallback to req ctx id usr-456, got %q (ok=%v)", id, ok)
	}

	role, ok := GetUserRole(c)
	if !ok || role != "user" {
		t.Fatalf("expected fallback to req ctx role user, got %q (ok=%v)", role, ok)
	}
}

func TestAuthContext_Empty(t *testing.T) {
	t.Parallel()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	if _, ok := GetUserID(c); ok {
		t.Fatal("expected GetUserID to return false on empty context")
	}
	if _, ok := GetUserRole(c); ok {
		t.Fatal("expected GetUserRole to return false on empty context")
	}
}


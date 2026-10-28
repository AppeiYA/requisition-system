package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParse_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	q, ok := Parse(c)
	if !ok {
		t.Fatal("expected ok")
	}
	if q.Limit != DefaultLimit || q.Offset != 0 {
		t.Fatalf("got %+v", q)
	}
}

func TestParse_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/test?limit=10&offset=20", nil)

	q, ok := Parse(c)
	if !ok {
		t.Fatal("expected ok")
	}
	if q.Limit != 10 || q.Offset != 20 {
		t.Fatalf("got %+v", q)
	}
}

func TestParse_InvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/test?limit=101", nil)

	if _, ok := Parse(c); ok {
		t.Fatal("expected not ok")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestParse_InvalidOffset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/test?offset=-1", nil)

	if _, ok := Parse(c); ok {
		t.Fatal("expected not ok")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestClamp(t *testing.T) {
	l, o := Clamp(0, -5)
	if l != DefaultLimit || o != 0 {
		t.Fatalf("got %d %d", l, o)
	}
	l, o = Clamp(200, 10)
	if l != MaxLimit || o != 10 {
		t.Fatalf("got %d %d", l, o)
	}
}

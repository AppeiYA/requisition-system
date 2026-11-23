package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/authcontext"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGinLogger_NormalRequest(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Options{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	router := gin.New()
	router.Use(GinLogger(l))
	router.GET("/api/v1/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON log: %v, raw: %s", err, buf.String())
	}

	httpData, ok := parsed["http"].(map[string]any)
	if !ok {
		t.Fatalf("missing http group: %v", parsed)
	}

	if httpData["method"] != "GET" {
		t.Errorf("method = %v, want GET", httpData["method"])
	}
	if httpData["route"] != "/api/v1/test" {
		t.Errorf("route = %v, want /api/v1/test", httpData["route"])
	}
	if httpData["status_code"] != float64(200) {
		t.Errorf("status_code = %v, want 200", httpData["status_code"])
	}
}

func TestGinLogger_WithAuthContext(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Options{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	router := gin.New()
	router.Use(func(c *gin.Context) {
		authcontext.SetUser(c, "usr-999", "admin")
		c.Next()
	})
	router.Use(GinLogger(l))
	router.GET("/api/v1/admin/action", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/action", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON log: %v", err)
	}

	userData, ok := parsed["user"].(map[string]any)
	if !ok {
		t.Fatalf("missing user group: %v", parsed)
	}

	if userData["id"] != "usr-999" {
		t.Errorf("user.id = %v, want usr-999", userData["id"])
	}
	if userData["role"] != "admin" {
		t.Errorf("user.role = %v, want admin", userData["role"])
	}
}

func TestGinLogger_HealthzDebugLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Options{
		Level:  "info", // Info level should drop DEBUG logs
		Format: "json",
		Output: buf,
	})

	router := gin.New()
	router.Use(GinLogger(l))
	router.GET("/api/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if buf.Len() > 0 {
		t.Errorf("expected healthz to be silenced at INFO level, got %s", buf.String())
	}
}

func TestGinRecovery_PanicsHandledAndLogged(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Options{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	router := gin.New()
	router.Use(GinRecovery(l))
	router.GET("/api/v1/panic", func(c *gin.Context) {
		panic("unexpected nil pointer")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON log: %v, raw: %s", err, buf.String())
	}

	if parsed["level"] != "ERROR" {
		t.Errorf("level = %v, want ERROR", parsed["level"])
	}
	if parsed["msg"] != "panic recovered in http handler" {
		t.Errorf("msg = %v, want 'panic recovered in http handler'", parsed["msg"])
	}
	if parsed["panic"] != "unexpected nil pointer" {
		t.Errorf("panic = %v, want 'unexpected nil pointer'", parsed["panic"])
	}
	if parsed["error.stack"] == "" {
		t.Errorf("expected stack trace in error.stack")
	}
}

func TestResponse_ErrorFromLogsInternalError(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Options{
		Level:  "info",
		Format: "json",
		Output: buf,
	})
	SetDefault(l)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/fail", nil)

	response.ErrorFrom(c, errors.New("db connection reset"))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON log: %v, raw: %s", err, buf.String())
	}

	if parsed["level"] != "ERROR" {
		t.Errorf("level = %v, want ERROR", parsed["level"])
	}
	if parsed["error.message"] != "db connection reset" {
		t.Errorf("error.message = %v, want 'db connection reset'", parsed["error.message"])
	}
}

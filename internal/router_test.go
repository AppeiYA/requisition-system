package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRenderScalarReference ensures the embedded OpenAPI specification parses
// and that the Scalar API reference page can be rendered from it.
func TestRenderScalarReference(t *testing.T) {
	html, err := renderScalarReference()
	if err != nil {
		t.Fatalf("renderScalarReference() error = %v", err)
	}

	if !strings.Contains(html, `id="api-reference"`) {
		t.Error("rendered HTML does not contain the Scalar api-reference element")
	}

	if !strings.Contains(html, "Requisition System API") {
		t.Error("rendered HTML does not contain the page title")
	}
}

func TestRequestBodyLimitSkipsOnlyUploadRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	largeBody := strings.Repeat("x", (2<<20)+1)

	newRouter := func() *gin.Engine {
		router := gin.New()
		router.Use(requestBodyLimit("/api/v1"))
		router.POST("/api/v1/attachments", func(c *gin.Context) {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
			c.Header("X-Read-Len", strconv.Itoa(len(body)))
			c.Status(http.StatusOK)
		})
		router.POST("/api/v1/uploads", func(c *gin.Context) {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
			c.Header("X-Read-Len", strconv.Itoa(len(body)))
			c.Status(http.StatusOK)
		})
		router.POST("/api/v1/requisitions", func(c *gin.Context) {
			_, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.Status(http.StatusRequestEntityTooLarge)
				return
			}
			c.Status(http.StatusOK)
		})
		return router
	}

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/attachments", strings.NewReader(largeBody))
	newRouter().ServeHTTP(resp, req)
	if got := resp.Header().Get("X-Read-Len"); got != strconv.Itoa(len(largeBody)) {
		t.Errorf("upload body length = %q, want %d (global limit must not apply)", got, len(largeBody))
	}

	resp = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/uploads", strings.NewReader(largeBody))
	newRouter().ServeHTTP(resp, req)
	if got := resp.Header().Get("X-Read-Len"); got != strconv.Itoa(len(largeBody)) {
		t.Errorf("staged upload body length = %q, want %d (global limit must not apply)", got, len(largeBody))
	}

	resp = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/requisitions", strings.NewReader(largeBody))
	newRouter().ServeHTTP(resp, req)
	if resp.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("JSON route status = %d, want 413", resp.Code)
	}
}

func TestAttachmentUploadPath(t *testing.T) {
	tests := map[string]string{
		"/api/v1":                           "/api/v1/attachments",
		"/api/v1/":                          "/api/v1/attachments",
		"http://localhost:8080/api/v1":      "/api/v1/attachments",
		"https://example.com/api/v1/upload": "/api/v1/upload/attachments",
	}
	for input, want := range tests {
		if got := attachmentUploadPath(input); got != want {
			t.Errorf("attachmentUploadPath(%q) = %q, want %q", input, got, want)
		}
	}
}

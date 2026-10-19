package app

import (
	"strings"
	"testing"
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

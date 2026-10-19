package generator

import (
	"strings"
	"testing"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

var _ identityout.IDGenerator = (*IDGenerator)(nil)

func Test_IDGenerator_GenerateID(t *testing.T) {
	gen := &IDGenerator{}
	prefix := "USR"

	id, err := gen.GenerateID(prefix)
	if err != nil {
		t.Fatalf("unexpected error generating ID: %v", err)
	}

	expectedPrefix := prefix + "-"
	if !strings.HasPrefix(id, expectedPrefix) {
		t.Errorf("expected ID to start with %s, got: %s", expectedPrefix, id)
	}

	suffix := strings.TrimPrefix(id, expectedPrefix)
	if len(suffix) != 8 {
		t.Errorf("expected suffix length of 8, got %d (id: %s)", len(suffix), id)
	}
}

package storage

import (
	"testing"

	"github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/storage/local"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
)

// TestLocalContract keeps the shared suite executable from the parent package
// as well; the local package's own test calls RunContract directly.
func TestLocalContract(t *testing.T) {
	RunContract(t, func(t *testing.T) attachmentout.FileStorage {
		return local.New(t.TempDir())
	})
}

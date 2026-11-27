package local_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/storage"
	"github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/storage/local"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
)

func TestLocalStorageContract(t *testing.T) {
	storage.RunContract(t, func(t *testing.T) attachmentout.FileStorage {
		return local.New(t.TempDir())
	})
}

func TestNewCreatesPrivateRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "attachments")
	_ = local.New(root)

	info, err := os.Stat(root)
	if err != nil {
		t.Fatalf("stat storage root: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("storage root is not a directory")
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Fatalf("storage root mode = %o, want 700", got)
	}
}

func TestPutUsesPrivateFilesAndDirectories(t *testing.T) {
	root := t.TempDir()
	store := local.New(root)
	const key = "requisition/req-1/att-1"

	if _, err := store.Put(context.Background(), key, bytes.NewReader([]byte("private")), attachmentout.PutOptions{
		ContentType: "application/pdf",
	}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	stream, _, err := store.Open(context.Background(), key)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer stream.Close()
	if _, ok := stream.(*os.File); !ok {
		t.Fatalf("Open reader type = %T, want *os.File for range support", stream)
	}

	objectPath := filepath.Join(root, filepath.FromSlash(key))
	for _, path := range []string{objectPath, objectPath + ".meta"} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", filepath.Base(path), err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Errorf("%s mode = %o, want 600", filepath.Base(path), got)
		}
	}
	info, err := os.Stat(filepath.Dir(objectPath))
	if err != nil {
		t.Fatalf("stat object directory: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Errorf("object directory mode = %o, want 700", got)
	}
}

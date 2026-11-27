// Package storage holds the shared behavioural contract for FileStorage
// adapters. The suite lives in a normal Go file so other packages can import
// it from their tests; a *_test.go file is not visible to importing packages.
package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"testing"

	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

const (
	contractKey      = "requisition/req-0123456789abcdef/att-0123456789abcdef"
	contractItemKey  = "requisition_item/reqitem-0123456789abcdef/att-0123456789abcdef"
	contractStageKey = "staging/upl-0123456789abcdef"
)

// RunContract exercises the FileStorage port against a fresh adapter instance.
// Every adapter must satisfy it unchanged.
func RunContract(t *testing.T, newStorage func(t *testing.T) attachmentout.FileStorage) {
	t.Helper()

	t.Run("PutStatOpenDelete", func(t *testing.T) {
		store := newStorage(t)
		ctx := context.Background()
		payload := []byte("attachment bytes for the round trip")

		putInfo, err := store.Put(ctx, contractKey, bytes.NewReader(payload), attachmentout.PutOptions{
			ContentType: "application/pdf",
			Size:        int64(len(payload) + 42), // deliberately wrong
		})
		if err != nil {
			t.Fatalf("Put: %v", err)
		}
		assertInfo(t, putInfo, contractKey, int64(len(payload)), "application/pdf", sha256Hex(payload))

		statInfo, err := store.Stat(ctx, contractKey)
		if err != nil {
			t.Fatalf("Stat after Put: %v", err)
		}
		assertInfo(t, statInfo, contractKey, int64(len(payload)), "application/pdf", sha256Hex(payload))

		stream, openInfo, err := store.Open(ctx, contractKey)
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		defer stream.Close()
		assertInfo(t, openInfo, contractKey, int64(len(payload)), "application/pdf", sha256Hex(payload))

		got, err := io.ReadAll(stream)
		if err != nil {
			t.Fatalf("read opened object: %v", err)
		}
		if !bytes.Equal(got, payload) {
			t.Fatalf("opened bytes = %q, want %q", got, payload)
		}
		if seeker, ok := stream.(io.ReadSeeker); ok {
			if _, err := seeker.Seek(0, io.SeekStart); err != nil {
				t.Fatalf("seek opened object: %v", err)
			}
		}

		if err := store.Delete(ctx, contractKey); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		assertMissing(t, store, contractKey)
	})

	t.Run("SizeAndChecksumComeFromBytes", func(t *testing.T) {
		store := newStorage(t)
		ctx := context.Background()
		payload := []byte("name,role\nalice,admin\n")

		info, err := store.Put(ctx, contractItemKey, bytes.NewReader(payload), attachmentout.PutOptions{
			ContentType: "text/csv",
			Size:        1,
		})
		if err != nil {
			t.Fatalf("Put: %v", err)
		}
		if info.Size != int64(len(payload)) {
			t.Fatalf("Put size = %d, want %d", info.Size, len(payload))
		}
		if want := sha256Hex(payload); info.Checksum != want {
			t.Fatalf("Put checksum = %q, want %q", info.Checksum, want)
		}

		statInfo, err := store.Stat(ctx, contractItemKey)
		if err != nil {
			t.Fatalf("Stat: %v", err)
		}
		if statInfo.Size != int64(len(payload)) || statInfo.Checksum != sha256Hex(payload) {
			t.Fatalf("Stat size/checksum = %d/%q, want %d/%q",
				statInfo.Size, statInfo.Checksum, len(payload), sha256Hex(payload))
		}
	})

	t.Run("PutComputesChecksumFromWrittenBytes", func(t *testing.T) {
		store := newStorage(t)
		ctx := context.Background()
		payload := []byte("stored with a computed checksum")

		putInfo, err := store.Put(ctx, contractKey, bytes.NewReader(payload), attachmentout.PutOptions{
			ContentType: "image/png",
		})
		if err != nil {
			t.Fatalf("Put: %v", err)
		}
		want := sha256Hex(payload)
		if putInfo.Checksum != want {
			t.Fatalf("Put checksum = %q, want %q", putInfo.Checksum, want)
		}
		info, err := store.Stat(ctx, contractKey)
		if err != nil {
			t.Fatalf("Stat: %v", err)
		}
		if info.Checksum != want {
			t.Fatalf("Stat checksum = %q, want %q", info.Checksum, want)
		}
	})

	t.Run("PutFailureLeavesNothingStored", func(t *testing.T) {
		store := newStorage(t)
		ctx := context.Background()

		if _, err := store.Put(ctx, contractKey, &failingReader{}, attachmentout.PutOptions{
			ContentType: "application/pdf",
		}); err == nil {
			t.Fatal("Put with a failing reader unexpectedly succeeded")
		}
		assertMissing(t, store, contractKey)
	})

	t.Run("MissingOpen", func(t *testing.T) {
		store := newStorage(t)
		stream, _, err := store.Open(context.Background(), contractKey)
		if stream != nil {
			stream.Close()
		}
		if !attachmentout.IsObjectNotFound(err) {
			t.Fatalf("Open missing object error = %v, want ErrObjectNotFound", err)
		}
	})

	t.Run("DeleteMissing", func(t *testing.T) {
		store := newStorage(t)
		if err := store.Delete(context.Background(), contractKey); err != nil {
			t.Fatalf("Delete missing object: %v", err)
		}
	})

	t.Run("MoveStagingToPermanentIsIdempotent", func(t *testing.T) {
		store := newStorage(t)
		ctx := context.Background()
		payload := []byte("staged quote bytes")

		if _, err := store.Put(ctx, contractStageKey, bytes.NewReader(payload), attachmentout.PutOptions{
			ContentType: "application/pdf",
		}); err != nil {
			t.Fatalf("Put staging: %v", err)
		}
		moved, err := store.Move(ctx, contractStageKey, contractKey, sha256Hex(payload))
		if err != nil {
			t.Fatalf("Move: %v", err)
		}
		assertInfo(t, moved, contractKey, int64(len(payload)), "application/pdf", sha256Hex(payload))
		assertMissing(t, store, contractStageKey)

		// A worker that crashes after the rename and retries sees the object
		// already at its destination and completes the operation.
		retried, err := store.Move(ctx, contractStageKey, contractKey, sha256Hex(payload))
		if err != nil {
			t.Fatalf("Move retry: %v", err)
		}
		assertInfo(t, retried, contractKey, int64(len(payload)), "application/pdf", sha256Hex(payload))
	})

	t.Run("MoveChecksChecksumAndMissingObjects", func(t *testing.T) {
		store := newStorage(t)
		ctx := context.Background()
		payload := []byte("integrity checked bytes")

		if _, err := store.Move(ctx, contractStageKey, contractKey, ""); !attachmentout.IsObjectNotFound(err) {
			t.Fatalf("Move missing objects error = %v, want ErrObjectNotFound", err)
		}

		if _, err := store.Put(ctx, contractStageKey, bytes.NewReader(payload), attachmentout.PutOptions{
			ContentType: "application/pdf",
		}); err != nil {
			t.Fatalf("Put staging: %v", err)
		}
		_, err := store.Move(ctx, contractStageKey, contractKey, sha256Hex([]byte("different")))
		if !errors.Is(err, attachmentout.ErrChecksumMismatch) {
			t.Fatalf("Move checksum mismatch error = %v, want ErrChecksumMismatch", err)
		}
		assertMissing(t, store, contractKey)
		stream, _, err := store.Open(ctx, contractStageKey)
		if err != nil {
			t.Fatalf("source was removed after a checksum mismatch: %v", err)
		}
		stream.Close()
	})

	t.Run("InvalidAndTraversalKeys", func(t *testing.T) {
		keys := []string{
			"",
			"requisition/req-0123456789abcdef",
			"requisition/req-0123456789abcdef/att-0123456789abcdef/extra",
			"/requisition/req-0123456789abcdef/att-0123456789abcdef",
			"requisition//att-0123456789abcdef",
			"requisition/req-0123456789abcdef/",
			"unknown/req-0123456789abcdef/att-0123456789abcdef",
			"requisition/reqitem-0123456789abcdef/att-0123456789abcdef",
			"requisition_item/req-0123456789abcdef/att-0123456789abcdef",
			"requisition/req-0123456789abcdef/attachment-1",
			"requisition/req-../att-0123456789abcdef",
			"requisition/req-0123456789abcdef/..",
			"requisition/req-0123456789abcdef/../att-0123456789abcdef",
			`requisition\req-0123456789abcdef\att-0123456789abcdef`,
			"staging",
			"staging/upl-0123456789abcdef/extra",
			"staging/att-0123456789abcdef",
		}
		for _, key := range keys {
			t.Run(key, func(t *testing.T) {
				store := newStorage(t)
				ctx := context.Background()

				_, err := store.Put(ctx, key, bytes.NewReader([]byte("x")), attachmentout.PutOptions{})
				assertBadRequest(t, err)
				stream, _, err := store.Open(ctx, key)
				if stream != nil {
					stream.Close()
				}
				assertBadRequest(t, err)
				_, err = store.Stat(ctx, key)
				assertBadRequest(t, err)
				assertBadRequest(t, store.Delete(ctx, key))
				_, err = store.Move(ctx, key, contractKey, "")
				assertBadRequest(t, err)
				_, err = store.Move(ctx, contractStageKey, key, "")
				assertBadRequest(t, err)
			})
		}
	})
}

func assertInfo(t *testing.T, info attachmentout.ObjectInfo, key string, size int64, contentType, checksum string) {
	t.Helper()
	if info.Key != key {
		t.Errorf("Key = %q, want %q", info.Key, key)
	}
	if info.Size != size {
		t.Errorf("Size = %d, want %d", info.Size, size)
	}
	if info.ContentType != contentType {
		t.Errorf("ContentType = %q, want %q", info.ContentType, contentType)
	}
	if info.Checksum != checksum {
		t.Errorf("Checksum = %q, want %q", info.Checksum, checksum)
	}
	if info.ModifiedAt.IsZero() {
		t.Error("ModifiedAt is zero")
	}
}

func assertMissing(t *testing.T, store attachmentout.FileStorage, key string) {
	t.Helper()
	ctx := context.Background()
	if _, err := store.Stat(ctx, key); !attachmentout.IsObjectNotFound(err) {
		t.Errorf("Stat after delete error = %v, want ErrObjectNotFound", err)
	}
	stream, _, err := store.Open(ctx, key)
	if stream != nil {
		stream.Close()
	}
	if !attachmentout.IsObjectNotFound(err) {
		t.Errorf("Open after delete error = %v, want ErrObjectNotFound", err)
	}
}

func assertBadRequest(t *testing.T, err error) {
	t.Helper()
	var apiErr *customerrors.CustomError
	if !errors.As(err, &apiErr) || apiErr.Code != http.StatusBadRequest {
		t.Errorf("error = %v, want HTTP 400 bad request", err)
	}
}

func sha256Hex(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

type failingReader struct {
	sent bool
}

func (r *failingReader) Read(p []byte) (int, error) {
	if r.sent {
		return 0, errors.New("simulated read failure")
	}
	r.sent = true
	return copy(p, []byte("partial")), nil
}

package attachmentout

import (
	"context"
	"errors"
	"io"
	"time"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrObjectNotFound     = customerrors.NewNotFoundError("stored object not found")
	ErrStorageUnavailable = customerrors.NewInternalServerError("file storage unavailable")
	ErrChecksumMismatch   = customerrors.NewConflictError("stored object checksum mismatch")
)

// ObjectInfo describes one stored object. It contains nothing driver specific.
type ObjectInfo struct {
	Key         string
	Size        int64
	ContentType string
	Checksum    string // lowercase hex SHA-256 of the stored bytes
	ModifiedAt  time.Time
}

type PutOptions struct {
	ContentType string
	Size        int64
}

// FileStorage is the only way the application reaches file bytes.
type FileStorage interface {
	// Put stores r under key. A non-nil error means nothing was stored.
	Put(ctx context.Context, key string, r io.Reader, opts PutOptions) (ObjectInfo, error)

	// Open returns a seekable stream and metadata. It must return
	// ErrObjectNotFound for an unknown key.
	Open(ctx context.Context, key string) (io.ReadSeekCloser, ObjectInfo, error)

	// Stat returns metadata without reading the object.
	Stat(ctx context.Context, key string) (ObjectInfo, error)

	// Delete removes the object. Deleting a missing key is not an error.
	Delete(ctx context.Context, key string) error

	// Move relocates an object from a staging key to its permanent key. It is
	// idempotent: if from is missing and to exists with the expected checksum,
	// the move is treated as already done. A mismatch returns
	// ErrChecksumMismatch. Both missing returns ErrObjectNotFound.
	Move(ctx context.Context, from, to, expectedChecksum string) (ObjectInfo, error)
}

// SignedURLProvider is implemented only by adapters that can hand out a
// time-limited URL (S3, MinIO, any CDN-backed store). A nil provider means
// downloads are streamed through the API.
type SignedURLProvider interface {
	SignedURL(ctx context.Context, key string, ttl time.Duration, downloadName string) (string, time.Time, error)
}

func IsObjectNotFound(err error) bool {
	return errors.Is(err, ErrObjectNotFound)
}

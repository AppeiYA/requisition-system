package s3_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/storage"
	s3adapter "github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/storage/s3"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
)

const fakeBucket = "attachments"

type fakeObject struct {
	body        []byte
	contentType string
	checksum    string
	modified    time.Time
}

type fakeS3 struct {
	mu      sync.Mutex
	objects map[string]fakeObject
}

func newFakeS3() *fakeS3 {
	return &fakeS3{objects: make(map[string]fakeObject)}
}

func (f *fakeS3) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	prefix := "/" + fakeBucket + "/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		http.NotFound(w, r)
		return
	}
	key := strings.TrimPrefix(r.URL.Path, prefix)

	f.mu.Lock()
	defer f.mu.Unlock()

	switch r.Method {
	case http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}
		f.objects[key] = fakeObject{
			body:        body,
			contentType: r.Header.Get("Content-Type"),
			checksum:    r.Header.Get("X-Amz-Meta-Attachment-Checksum"),
			modified:    time.Now().UTC(),
		}
		w.WriteHeader(http.StatusOK)
	case http.MethodHead, http.MethodGet:
		object, ok := f.objects[key]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", object.contentType)
		w.Header().Set("Content-Length", strconv.Itoa(len(object.body)))
		w.Header().Set("Last-Modified", object.modified.Format(http.TimeFormat))
		if object.checksum != "" {
			w.Header().Set("X-Amz-Meta-Attachment-Checksum", object.checksum)
		}
		if disposition := r.URL.Query().Get("response-content-disposition"); disposition != "" {
			w.Header().Set("Content-Disposition", disposition)
		}
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodGet {
			_, _ = w.Write(object.body)
		}
	case http.MethodDelete:
		delete(f.objects, key)
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "unsupported method", http.StatusMethodNotAllowed)
	}
}

func newTestStorage(t *testing.T) attachmentout.FileStorage {
	t.Helper()
	server := httptest.NewServer(newFakeS3())
	t.Cleanup(server.Close)

	store, err := s3adapter.New(s3adapter.Options{
		Bucket:          fakeBucket,
		Region:          "us-east-1",
		Endpoint:        server.URL,
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
		ForcePathStyle:  true,
		UseSSL:          false,
	})
	require.NoError(t, err)
	return store
}

func TestS3StorageContract(t *testing.T) {
	storage.RunContract(t, newTestStorage)
}

func TestS3SignedURL(t *testing.T) {
	server := httptest.NewServer(newFakeS3())
	t.Cleanup(server.Close)
	store, err := s3adapter.New(s3adapter.Options{
		Bucket:          fakeBucket,
		Region:          "us-east-1",
		Endpoint:        server.URL,
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
		ForcePathStyle:  true,
	})
	require.NoError(t, err)

	ctx := context.Background()
	const key = "requisition/req-0123456789abcdef/att-0123456789abcdef"
	payload := []byte("signed attachment")
	_, err = store.Put(ctx, key, bytes.NewReader(payload), attachmentout.PutOptions{
		ContentType: "application/pdf",
	})
	require.NoError(t, err)

	location, expiresAt, err := store.SignedURL(ctx, key, 5*time.Minute, "signed report.pdf")
	require.NoError(t, err)
	require.Contains(t, location, "X-Amz-Signature=")
	require.WithinDuration(t, time.Now().Add(5*time.Minute), expiresAt, 10*time.Second)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
	require.NoError(t, err)
	response, err := server.Client().Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Contains(t, response.Header.Get("Content-Disposition"), `filename="signed report.pdf"`)
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, payload, body)
}

func TestS3NewValidatesOptions(t *testing.T) {
	tests := []struct {
		name string
		opts s3adapter.Options
	}{
		{name: "missing bucket", opts: s3adapter.Options{}},
		{
			name: "half credentials",
			opts: s3adapter.Options{Bucket: fakeBucket, AccessKeyID: "key"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s3adapter.New(tt.opts)
			require.Error(t, err)
		})
	}
}

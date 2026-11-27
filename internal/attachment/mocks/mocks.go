// Package attachmentmocks provides hand-written test doubles for the
// attachment outbound ports. Every method has a functional field plus a
// harmless default, so a mock built with a constructor is immediately usable.
package attachmentmocks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
)

const (
	DefaultAttachmentID = "att-1234"
	DefaultRefID        = "req-1234"
	DefaultRefName      = "requisition"
	DefaultStorageKey   = "requisition/req-1234/att-1234"
	DefaultFileName     = "quote.pdf"
	DefaultActorID      = "usr-5678"
	DefaultActorRole    = "user"
	DefaultFileSize     = int64(1024)
	DefaultStagedID     = "upl-1234"
	DefaultStagingPath  = "staging/upl-1234"
)

// ValidAttachment returns a minimal, domain-valid attachment for happy paths.
func ValidAttachment() *domain.Attachment {
	ref, _ := domain.NewAttachmentRef(domain.RefNameRequisition, DefaultRefID)
	name, _ := domain.NewAttachmentFileName(DefaultFileName)
	attachment, _ := domain.NewAttachment(
		DefaultAttachmentID,
		ref,
		DefaultStorageKey,
		name,
		domain.ContentTypePDF,
		DefaultFileSize,
		strings.Repeat("a", domain.ChecksumLength),
		DefaultActorID,
	)
	return attachment
}

// ValidStagedUpload returns a minimal, domain-valid pending upload for happy paths.
func ValidStagedUpload() *domain.StagedUpload {
	name, _ := domain.NewAttachmentFileName(DefaultFileName)
	upload, _ := domain.NewStagedUpload(
		DefaultStagedID,
		DefaultActorID,
		DefaultStagingPath,
		name,
		domain.ContentTypePDF,
		DefaultFileSize,
		strings.Repeat("a", domain.ChecksumLength),
		24*time.Hour,
		time.Now(),
	)
	return upload
}

// ValidReconciliationEvent returns a pending event for the default attachment.
func ValidReconciliationEvent() *domain.ReconciliationEvent {
	event, _ := domain.NewReconciliationEvent(
		"aevt-1234",
		DefaultAttachmentID,
		DefaultStagedID,
		DefaultStagingPath,
		DefaultStorageKey,
		strings.Repeat("a", domain.ChecksumLength),
		time.Now().UTC(),
	)
	return event
}

// MockStagedUploadRepository is a test double for attachmentout.StagedUploadRepository.
type MockStagedUploadRepository struct {
	SaveFunc               func(ctx context.Context, upload *domain.StagedUpload) error
	GetByIDFunc            func(ctx context.Context, id string) (*domain.StagedUpload, error)
	ClaimByIDsFunc         func(ctx context.Context, ids []string, actorID string, now time.Time) ([]*domain.StagedUpload, error)
	CountPendingByUserFunc func(ctx context.Context, actorID string, now time.Time) (int, int64, error)
	ExpirePendingFunc      func(ctx context.Context, now time.Time, limit int) ([]*domain.StagedUpload, error)
	CountPendingAllFunc    func(ctx context.Context, now time.Time) (int64, error)
}

func NewMockStagedUploadRepository() *MockStagedUploadRepository {
	return &MockStagedUploadRepository{
		SaveFunc:    func(context.Context, *domain.StagedUpload) error { return nil },
		GetByIDFunc: func(context.Context, string) (*domain.StagedUpload, error) { return ValidStagedUpload(), nil },
		ClaimByIDsFunc: func(context.Context, []string, string, time.Time) ([]*domain.StagedUpload, error) {
			return []*domain.StagedUpload{ValidStagedUpload()}, nil
		},
		CountPendingByUserFunc: func(context.Context, string, time.Time) (int, int64, error) { return 0, 0, nil },
		ExpirePendingFunc:      func(context.Context, time.Time, int) ([]*domain.StagedUpload, error) { return nil, nil },
		CountPendingAllFunc:    func(context.Context, time.Time) (int64, error) { return 0, nil },
	}
}

func (m *MockStagedUploadRepository) Save(ctx context.Context, upload *domain.StagedUpload) error {
	if m.SaveFunc == nil {
		return nil
	}
	return m.SaveFunc(ctx, upload)
}

func (m *MockStagedUploadRepository) GetByID(ctx context.Context, id string) (*domain.StagedUpload, error) {
	if m.GetByIDFunc == nil {
		return ValidStagedUpload(), nil
	}
	return m.GetByIDFunc(ctx, id)
}

func (m *MockStagedUploadRepository) ClaimByIDs(
	ctx context.Context,
	ids []string,
	actorID string,
	now time.Time,
) ([]*domain.StagedUpload, error) {
	if m.ClaimByIDsFunc == nil {
		return []*domain.StagedUpload{ValidStagedUpload()}, nil
	}
	return m.ClaimByIDsFunc(ctx, ids, actorID, now)
}

func (m *MockStagedUploadRepository) CountPendingByUser(ctx context.Context, actorID string, now time.Time) (int, int64, error) {
	if m.CountPendingByUserFunc == nil {
		return 0, 0, nil
	}
	return m.CountPendingByUserFunc(ctx, actorID, now)
}

func (m *MockStagedUploadRepository) ExpirePending(ctx context.Context, now time.Time, limit int) ([]*domain.StagedUpload, error) {
	if m.ExpirePendingFunc == nil {
		return nil, nil
	}
	return m.ExpirePendingFunc(ctx, now, limit)
}

func (m *MockStagedUploadRepository) CountPendingAll(ctx context.Context, now time.Time) (int64, error) {
	if m.CountPendingAllFunc == nil {
		return 0, nil
	}
	return m.CountPendingAllFunc(ctx, now)
}

var _ attachmentout.StagedUploadRepository = (*MockStagedUploadRepository)(nil)

type nopReadSeekCloser struct {
	*bytes.Reader
}

func (nopReadSeekCloser) Close() error { return nil }

// MockFileStorage is a test double for attachmentout.FileStorage.
type MockFileStorage struct {
	PutFunc    func(ctx context.Context, key string, r io.Reader, opts attachmentout.PutOptions) (attachmentout.ObjectInfo, error)
	OpenFunc   func(ctx context.Context, key string) (io.ReadSeekCloser, attachmentout.ObjectInfo, error)
	StatFunc   func(ctx context.Context, key string) (attachmentout.ObjectInfo, error)
	DeleteFunc func(ctx context.Context, key string) error
	MoveFunc   func(ctx context.Context, from, to, expectedChecksum string) (attachmentout.ObjectInfo, error)
}

func NewMockFileStorage() *MockFileStorage {
	return &MockFileStorage{
		PutFunc: func(_ context.Context, key string, r io.Reader, opts attachmentout.PutOptions) (attachmentout.ObjectInfo, error) {
			return mockPutObjectInfo(key, r, opts)
		},
		OpenFunc: func(_ context.Context, key string) (io.ReadSeekCloser, attachmentout.ObjectInfo, error) {
			return nopReadSeekCloser{bytes.NewReader(nil)}, attachmentout.ObjectInfo{Key: key}, nil
		},
		StatFunc: func(_ context.Context, key string) (attachmentout.ObjectInfo, error) {
			return attachmentout.ObjectInfo{Key: key}, nil
		},
		DeleteFunc: func(context.Context, string) error { return nil },
		MoveFunc: func(_ context.Context, _, to, expectedChecksum string) (attachmentout.ObjectInfo, error) {
			return attachmentout.ObjectInfo{Key: to, Checksum: expectedChecksum}, nil
		},
	}
}

func mockPutObjectInfo(key string, r io.Reader, opts attachmentout.PutOptions) (attachmentout.ObjectInfo, error) {
	h := sha256.New()
	var n int64
	if r != nil {
		var err error
		n, err = io.Copy(h, r)
		if err != nil {
			return attachmentout.ObjectInfo{}, err
		}
	}
	size := n
	if size == 0 {
		size = opts.Size
	}
	return attachmentout.ObjectInfo{
		Key:         key,
		Size:        size,
		ContentType: opts.ContentType,
		Checksum:    hex.EncodeToString(h.Sum(nil)),
	}, nil
}

func (m *MockFileStorage) Put(ctx context.Context, key string, r io.Reader, opts attachmentout.PutOptions) (attachmentout.ObjectInfo, error) {
	if m.PutFunc == nil {
		return mockPutObjectInfo(key, r, opts)
	}
	return m.PutFunc(ctx, key, r, opts)
}

func (m *MockFileStorage) Open(ctx context.Context, key string) (io.ReadSeekCloser, attachmentout.ObjectInfo, error) {
	if m.OpenFunc == nil {
		return nopReadSeekCloser{bytes.NewReader(nil)}, attachmentout.ObjectInfo{Key: key}, nil
	}
	return m.OpenFunc(ctx, key)
}

func (m *MockFileStorage) Stat(ctx context.Context, key string) (attachmentout.ObjectInfo, error) {
	if m.StatFunc == nil {
		return attachmentout.ObjectInfo{Key: key}, nil
	}
	return m.StatFunc(ctx, key)
}

func (m *MockFileStorage) Delete(ctx context.Context, key string) error {
	if m.DeleteFunc == nil {
		return nil
	}
	return m.DeleteFunc(ctx, key)
}

func (m *MockFileStorage) Move(ctx context.Context, from, to, expectedChecksum string) (attachmentout.ObjectInfo, error) {
	if m.MoveFunc == nil {
		return attachmentout.ObjectInfo{Key: to, Checksum: expectedChecksum}, nil
	}
	return m.MoveFunc(ctx, from, to, expectedChecksum)
}

var _ attachmentout.FileStorage = (*MockFileStorage)(nil)

// MockAttachmentRepository is a test double for attachmentout.AttachmentRepository.
type MockAttachmentRepository struct {
	SaveFunc                 func(ctx context.Context, attachment *domain.Attachment) error
	UpdateStatusFunc         func(ctx context.Context, id string, status domain.AttachmentStatus) error
	GetByIDFunc              func(ctx context.Context, id string) (*domain.Attachment, error)
	ListByRefFunc            func(ctx context.Context, refName, refID string, limit, offset int) ([]*domain.Attachment, int, error)
	CountByRefFunc           func(ctx context.Context, refName, refID string) (int, error)
	DeleteFunc               func(ctx context.Context, id string) error
	DeleteByRefFunc          func(ctx context.Context, refName, refID string) ([]string, error)
	ResetStuckProcessingFunc func(ctx context.Context, stuckBefore time.Time, limit int) ([]string, error)
}

func NewMockAttachmentRepository() *MockAttachmentRepository {
	return &MockAttachmentRepository{
		SaveFunc:         func(context.Context, *domain.Attachment) error { return nil },
		UpdateStatusFunc: func(context.Context, string, domain.AttachmentStatus) error { return nil },
		GetByIDFunc: func(context.Context, string) (*domain.Attachment, error) {
			return ValidAttachment(), nil
		},
		ListByRefFunc: func(context.Context, string, string, int, int) ([]*domain.Attachment, int, error) {
			return []*domain.Attachment{ValidAttachment()}, 1, nil
		},
		CountByRefFunc: func(context.Context, string, string) (int, error) { return 0, nil },
		DeleteFunc:     func(context.Context, string) error { return nil },
		DeleteByRefFunc: func(context.Context, string, string) ([]string, error) {
			return []string{DefaultStorageKey}, nil
		},
	}
}

func (m *MockAttachmentRepository) Save(ctx context.Context, attachment *domain.Attachment) error {
	if m.SaveFunc == nil {
		return nil
	}
	return m.SaveFunc(ctx, attachment)
}

func (m *MockAttachmentRepository) UpdateStatus(
	ctx context.Context,
	id string,
	status domain.AttachmentStatus,
) error {
	if m.UpdateStatusFunc == nil {
		return nil
	}
	return m.UpdateStatusFunc(ctx, id, status)
}

func (m *MockAttachmentRepository) GetByID(ctx context.Context, id string) (*domain.Attachment, error) {
	if m.GetByIDFunc == nil {
		return ValidAttachment(), nil
	}
	return m.GetByIDFunc(ctx, id)
}

func (m *MockAttachmentRepository) ListByRef(ctx context.Context, refName, refID string, limit, offset int) ([]*domain.Attachment, int, error) {
	if m.ListByRefFunc == nil {
		return []*domain.Attachment{ValidAttachment()}, 1, nil
	}
	return m.ListByRefFunc(ctx, refName, refID, limit, offset)
}

func (m *MockAttachmentRepository) CountByRef(ctx context.Context, refName, refID string) (int, error) {
	if m.CountByRefFunc == nil {
		return 0, nil
	}
	return m.CountByRefFunc(ctx, refName, refID)
}

func (m *MockAttachmentRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc == nil {
		return nil
	}
	return m.DeleteFunc(ctx, id)
}

func (m *MockAttachmentRepository) DeleteByRef(ctx context.Context, refName, refID string) ([]string, error) {
	if m.DeleteByRefFunc == nil {
		return []string{DefaultStorageKey}, nil
	}
	return m.DeleteByRefFunc(ctx, refName, refID)
}

func (m *MockAttachmentRepository) ResetStuckProcessing(ctx context.Context, stuckBefore time.Time, limit int) ([]string, error) {
	if m.ResetStuckProcessingFunc == nil {
		return nil, nil
	}
	return m.ResetStuckProcessingFunc(ctx, stuckBefore, limit)
}

var _ attachmentout.AttachmentRepository = (*MockAttachmentRepository)(nil)

// MockReconciliationOutboxRepository is a test double for the outbox writer.
type MockReconciliationOutboxRepository struct {
	SaveFunc              func(ctx context.Context, event *domain.ReconciliationEvent) error
	GetByIDFunc           func(ctx context.Context, id string) (*domain.ReconciliationEvent, error)
	GetByAttachmentIDFunc func(ctx context.Context, attachmentID string) (*domain.ReconciliationEvent, error)
	ClaimDueFunc          func(ctx context.Context, now time.Time) (*domain.ReconciliationEvent, error)
	MarkDoneFunc          func(ctx context.Context, id string, now time.Time) error
	RecordFailureFunc     func(
		ctx context.Context,
		id string,
		status domain.ReconciliationStatus,
		attempts int,
		nextAttemptAt time.Time,
		now time.Time,
	) error
	RequeueByAttachmentIDsFunc func(ctx context.Context, attachmentIDs []string, now time.Time) (int64, error)
	CountDepthFunc             func(ctx context.Context) (int64, int64, error)
}

func NewMockReconciliationOutboxRepository() *MockReconciliationOutboxRepository {
	return &MockReconciliationOutboxRepository{
		SaveFunc: func(context.Context, *domain.ReconciliationEvent) error { return nil },
		GetByIDFunc: func(context.Context, string) (*domain.ReconciliationEvent, error) {
			return ValidReconciliationEvent(), nil
		},
		GetByAttachmentIDFunc: func(context.Context, string) (*domain.ReconciliationEvent, error) {
			return nil, attachmentout.ErrReconciliationEventNotFound
		},
		ClaimDueFunc: func(context.Context, time.Time) (*domain.ReconciliationEvent, error) {
			return ValidReconciliationEvent(), nil
		},
		MarkDoneFunc: func(context.Context, string, time.Time) error { return nil },
		RecordFailureFunc: func(
			context.Context, string, domain.ReconciliationStatus, int, time.Time, time.Time,
		) error {
			return nil
		},
	}
}

func (m *MockReconciliationOutboxRepository) Save(
	ctx context.Context,
	event *domain.ReconciliationEvent,
) error {
	if m.SaveFunc == nil {
		return nil
	}
	return m.SaveFunc(ctx, event)
}

func (m *MockReconciliationOutboxRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.ReconciliationEvent, error) {
	if m.GetByIDFunc == nil {
		return ValidReconciliationEvent(), nil
	}
	return m.GetByIDFunc(ctx, id)
}

func (m *MockReconciliationOutboxRepository) GetByAttachmentID(
	ctx context.Context,
	attachmentID string,
) (*domain.ReconciliationEvent, error) {
	if m.GetByAttachmentIDFunc == nil {
		return nil, attachmentout.ErrReconciliationEventNotFound
	}
	return m.GetByAttachmentIDFunc(ctx, attachmentID)
}

func (m *MockReconciliationOutboxRepository) ClaimDue(
	ctx context.Context,
	now time.Time,
) (*domain.ReconciliationEvent, error) {
	if m.ClaimDueFunc == nil {
		return ValidReconciliationEvent(), nil
	}
	return m.ClaimDueFunc(ctx, now)
}

func (m *MockReconciliationOutboxRepository) MarkDone(
	ctx context.Context,
	id string,
	now time.Time,
) error {
	if m.MarkDoneFunc == nil {
		return nil
	}
	return m.MarkDoneFunc(ctx, id, now)
}

func (m *MockReconciliationOutboxRepository) RecordFailure(
	ctx context.Context,
	id string,
	status domain.ReconciliationStatus,
	attempts int,
	nextAttemptAt time.Time,
	now time.Time,
) error {
	if m.RecordFailureFunc == nil {
		return nil
	}
	return m.RecordFailureFunc(ctx, id, status, attempts, nextAttemptAt, now)
}

func (m *MockReconciliationOutboxRepository) RequeueByAttachmentIDs(ctx context.Context, attachmentIDs []string, now time.Time) (int64, error) {
	if m.RequeueByAttachmentIDsFunc == nil {
		return 0, nil
	}
	return m.RequeueByAttachmentIDsFunc(ctx, attachmentIDs, now)
}

func (m *MockReconciliationOutboxRepository) CountDepth(ctx context.Context) (int64, int64, error) {
	if m.CountDepthFunc == nil {
		return 0, 0, nil
	}
	return m.CountDepthFunc(ctx)
}

var _ attachmentout.ReconciliationOutboxRepository = (*MockReconciliationOutboxRepository)(nil)

// MockRefAccessChecker is a test double for attachmentout.RefAccessChecker.
type MockRefAccessChecker struct {
	CanAccessFunc func(ctx context.Context, refName, refID, actorID, actorRole string, action attachmentout.Action) error
}

func NewMockRefAccessChecker() *MockRefAccessChecker {
	return &MockRefAccessChecker{
		CanAccessFunc: func(context.Context, string, string, string, string, attachmentout.Action) error {
			return nil
		},
	}
}

func (m *MockRefAccessChecker) CanAccess(ctx context.Context, refName, refID, actorID, actorRole string, action attachmentout.Action) error {
	if m.CanAccessFunc == nil {
		return nil
	}
	return m.CanAccessFunc(ctx, refName, refID, actorID, actorRole, action)
}

var _ attachmentout.RefAccessChecker = (*MockRefAccessChecker)(nil)

// MockContentTypeDetector is a test double for attachmentout.ContentTypeDetector.
type MockContentTypeDetector struct {
	DetectFunc func(head []byte, declared string) (domain.ContentType, error)
}

func NewMockContentTypeDetector() *MockContentTypeDetector {
	return &MockContentTypeDetector{
		DetectFunc: func([]byte, string) (domain.ContentType, error) {
			return domain.ContentTypePDF, nil
		},
	}
}

func (m *MockContentTypeDetector) Detect(head []byte, declared string) (domain.ContentType, error) {
	if m.DetectFunc == nil {
		return domain.ContentTypePDF, nil
	}
	return m.DetectFunc(head, declared)
}

var _ attachmentout.ContentTypeDetector = (*MockContentTypeDetector)(nil)

// MockIDGenerator is a test double for attachmentout.IDGenerator.
type MockIDGenerator struct {
	GenerateIDFunc func(prefix string) (string, error)
}

func NewMockIDGenerator() *MockIDGenerator {
	return &MockIDGenerator{
		GenerateIDFunc: func(prefix string) (string, error) {
			return prefix + "-1234", nil
		},
	}
}

func (m *MockIDGenerator) GenerateID(prefix string) (string, error) {
	if m.GenerateIDFunc == nil {
		return prefix + "-1234", nil
	}
	return m.GenerateIDFunc(prefix)
}

var _ attachmentout.IDGenerator = (*MockIDGenerator)(nil)

// MockTransactionManager is a test double for attachmentout.TransactionManager.
type MockTransactionManager struct {
	WithinTransactionFunc func(ctx context.Context, fn func(context.Context) error) error
}

func NewMockTransactionManager() *MockTransactionManager {
	return &MockTransactionManager{
		WithinTransactionFunc: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	}
}

func (m *MockTransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	if m.WithinTransactionFunc == nil {
		return fn(ctx)
	}
	return m.WithinTransactionFunc(ctx, fn)
}

var _ attachmentout.TransactionManager = (*MockTransactionManager)(nil)

// MockSignedURLProvider is a test double for attachmentout.SignedURLProvider.
type MockSignedURLProvider struct {
	SignedURLFunc func(ctx context.Context, key string, ttl time.Duration, downloadName string) (string, time.Time, error)
}

func NewMockSignedURLProvider() *MockSignedURLProvider {
	return &MockSignedURLProvider{
		SignedURLFunc: func(context.Context, string, time.Duration, string) (string, time.Time, error) {
			return "https://storage.example.test/signed", time.Now().Add(5 * time.Minute), nil
		},
	}
}

func (m *MockSignedURLProvider) SignedURL(ctx context.Context, key string, ttl time.Duration, downloadName string) (string, time.Time, error) {
	if m.SignedURLFunc == nil {
		return "https://storage.example.test/signed", time.Now().Add(5 * time.Minute), nil
	}
	return m.SignedURLFunc(ctx, key, ttl, downloadName)
}

var _ attachmentout.SignedURLProvider = (*MockSignedURLProvider)(nil)

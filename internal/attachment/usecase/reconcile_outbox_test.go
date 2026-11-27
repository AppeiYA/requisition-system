package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentmocks "github.com/AppeiYA/requisition-system/internal/attachment/mocks"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
)

var reconcileNow = time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC)

type reconcileFixture struct {
	uc           *ReconcileOutboxUsecase
	outbox       *attachmentmocks.MockReconciliationOutboxRepository
	attachments  *attachmentmocks.MockAttachmentRepository
	storage      *attachmentmocks.MockFileStorage
	metrics      *stubMetricsRecorder
	statusWrites []domain.AttachmentStatus
}

func pendingReconcileAttachment() *domain.Attachment {
	ref, _ := domain.NewAttachmentRef(domain.RefNameRequisition, attachmentmocks.DefaultRefID)
	name, _ := domain.NewAttachmentFileName(attachmentmocks.DefaultFileName)
	attachment, _ := domain.NewPendingAttachment(
		attachmentmocks.DefaultAttachmentID,
		ref,
		attachmentmocks.DefaultStagedID,
		attachmentmocks.DefaultStorageKey,
		name,
		domain.ContentTypePDF,
		attachmentmocks.DefaultFileSize,
		strings.Repeat("a", domain.ChecksumLength),
		attachmentmocks.DefaultActorID,
	)
	return attachment
}

func newReconcileFixture() *reconcileFixture {
	outbox := attachmentmocks.NewMockReconciliationOutboxRepository()
	attachments := attachmentmocks.NewMockAttachmentRepository()
	storage := attachmentmocks.NewMockFileStorage()
	metrics := &stubMetricsRecorder{}
	fixture := &reconcileFixture{outbox: outbox, attachments: attachments, storage: storage, metrics: metrics}

	attachments.GetByIDFunc = func(context.Context, string) (*domain.Attachment, error) {
		return pendingReconcileAttachment(), nil
	}
	attachments.UpdateStatusFunc = func(_ context.Context, _ string, status domain.AttachmentStatus) error {
		fixture.statusWrites = append(fixture.statusWrites, status)
		return nil
	}

	uc := NewReconcileOutboxUsecase(
		outbox, attachments, storage, attachmentmocks.NewMockTransactionManager(), metrics,
		1, 3, time.Minute,
	)
	uc.now = func() time.Time { return reconcileNow }
	fixture.uc = uc
	return fixture
}

func TestReconcileOutboxMovesBytesAndMarksEverythingReady(t *testing.T) {
	f := newReconcileFixture()
	var movedFrom, movedTo, movedChecksum string
	f.storage.MoveFunc = func(_ context.Context, from, to, checksum string) (attachmentout.ObjectInfo, error) {
		movedFrom, movedTo, movedChecksum = from, to, checksum
		return attachmentout.ObjectInfo{Key: to}, nil
	}
	var done bool
	f.outbox.MarkDoneFunc = func(_ context.Context, _ string, _ time.Time) error {
		done = true
		return nil
	}
	f.outbox.RecordFailureFunc = func(
		context.Context, string, domain.ReconciliationStatus, int, time.Time, time.Time,
	) error {
		t.Error("RecordFailure called on a successful reconciliation")
		return nil
	}

	if err := f.uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if movedFrom != attachmentmocks.DefaultStagingPath || movedTo != attachmentmocks.DefaultStorageKey {
		t.Errorf("Move() = %q -> %q", movedFrom, movedTo)
	}
	if movedChecksum != strings.Repeat("a", domain.ChecksumLength) {
		t.Errorf("Move() checksum = %q", movedChecksum)
	}
	if !done {
		t.Error("outbox event was not marked done")
	}
	if f.metrics.deadAdded != 0 {
		t.Errorf("dead-letter metric = %d, want 0", f.metrics.deadAdded)
	}
	want := []domain.AttachmentStatus{
		domain.AttachmentStatusProcessing,
		domain.AttachmentStatusReady,
	}
	if len(f.statusWrites) != len(want) {
		t.Fatalf("status writes = %v, want %v", f.statusWrites, want)
	}
	for i := range want {
		if f.statusWrites[i] != want[i] {
			t.Errorf("status write %d = %q, want %q", i, f.statusWrites[i], want[i])
		}
	}
}

func TestReconcileOutboxIsIdempotentWhenAttachmentIsReady(t *testing.T) {
	f := newReconcileFixture()
	f.attachments.GetByIDFunc = func(context.Context, string) (*domain.Attachment, error) {
		return attachmentmocks.ValidAttachment(), nil
	}
	f.storage.MoveFunc = func(context.Context, string, string, string) (attachmentout.ObjectInfo, error) {
		t.Error("Move() called for an already ready attachment")
		return attachmentout.ObjectInfo{}, nil
	}
	done := false
	f.outbox.MarkDoneFunc = func(context.Context, string, time.Time) error {
		done = true
		return nil
	}

	if err := f.uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !done {
		t.Error("outbox event was not marked done")
	}
	if len(f.statusWrites) != 0 {
		t.Errorf("status writes = %v, want none", f.statusWrites)
	}
}

func TestReconcileOutboxRetriesWithBackoffAndStaysUnready(t *testing.T) {
	f := newReconcileFixture()
	f.storage.MoveFunc = func(context.Context, string, string, string) (attachmentout.ObjectInfo, error) {
		return attachmentout.ObjectInfo{}, attachmentout.ErrStorageUnavailable
	}
	var status domain.ReconciliationStatus
	var attempts int
	var next time.Time
	f.outbox.RecordFailureFunc = func(
		_ context.Context,
		_ string,
		s domain.ReconciliationStatus,
		a int,
		n time.Time,
		_ time.Time,
	) error {
		status, attempts, next = s, a, n
		return nil
	}

	if err := f.uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if status != domain.ReconciliationStatusFailed || attempts != 1 {
		t.Fatalf("failure state = %q/%d, want failed/1", status, attempts)
	}
	minNext := reconcileNow.Add(time.Minute)
	maxNext := minNext.Add(15 * time.Second)
	if next.Before(minNext) || next.After(maxNext) {
		t.Errorf("next attempt = %v, want within [%v, %v]", next, minNext, maxNext)
	}
	if len(f.statusWrites) != 1 || f.statusWrites[0] != domain.AttachmentStatusProcessing {
		t.Errorf("status writes = %v, want only processing", f.statusWrites)
	}
	if f.metrics.deadAdded != 0 {
		t.Errorf("dead-letter metric = %d, want 0 for a retry", f.metrics.deadAdded)
	}
}

func TestReconcileOutboxMarksTerminalFailureAndAttachmentFailed(t *testing.T) {
	f := newReconcileFixture()
	f.uc = NewReconcileOutboxUsecase(
		f.outbox, f.attachments, f.storage, attachmentmocks.NewMockTransactionManager(), f.metrics,
		1, 1, time.Minute,
	)
	f.uc.now = func() time.Time { return reconcileNow }
	f.storage.MoveFunc = func(context.Context, string, string, string) (attachmentout.ObjectInfo, error) {
		return attachmentout.ObjectInfo{}, attachmentout.ErrChecksumMismatch
	}
	var status domain.ReconciliationStatus
	f.outbox.RecordFailureFunc = func(
		_ context.Context,
		_ string,
		s domain.ReconciliationStatus,
		_ int,
		_ time.Time,
		_ time.Time,
	) error {
		status = s
		return nil
	}

	if err := f.uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if status != domain.ReconciliationStatusDead {
		t.Errorf("failure status = %q, want dead", status)
	}
	if f.metrics.deadAdded != 1 {
		t.Errorf("dead-letter metric = %d, want 1", f.metrics.deadAdded)
	}
	want := []domain.AttachmentStatus{
		domain.AttachmentStatusProcessing,
		domain.AttachmentStatusFailed,
	}
	if len(f.statusWrites) != len(want) {
		t.Fatalf("status writes = %v, want %v", f.statusWrites, want)
	}
	for i := range want {
		if f.statusWrites[i] != want[i] {
			t.Errorf("status write %d = %q, want %q", i, f.statusWrites[i], want[i])
		}
	}
}

func TestReconcileOutboxTreatsMissingAttachmentAsTerminal(t *testing.T) {
	f := newReconcileFixture()
	f.attachments.GetByIDFunc = func(context.Context, string) (*domain.Attachment, error) {
		return nil, ErrAttachmentNotFound
	}
	f.storage.MoveFunc = func(context.Context, string, string, string) (attachmentout.ObjectInfo, error) {
		t.Error("Move() called without an attachment row")
		return attachmentout.ObjectInfo{}, nil
	}
	var status domain.ReconciliationStatus
	f.outbox.RecordFailureFunc = func(
		_ context.Context,
		_ string,
		s domain.ReconciliationStatus,
		_ int,
		_ time.Time,
		_ time.Time,
	) error {
		status = s
		return nil
	}

	if err := f.uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if status != domain.ReconciliationStatusDead {
		t.Errorf("failure status = %q, want dead", status)
	}
	if f.metrics.deadAdded != 1 {
		t.Errorf("dead-letter metric = %d, want 1", f.metrics.deadAdded)
	}
}

func TestReconcileOutboxDeadLettersAlreadyFailedAttachment(t *testing.T) {
	f := newReconcileFixture()
	f.attachments.GetByIDFunc = func(context.Context, string) (*domain.Attachment, error) {
		attachment := pendingReconcileAttachment()
		if err := attachment.MarkProcessing(); err != nil {
			return nil, err
		}
		if err := attachment.MarkFailed(); err != nil {
			return nil, err
		}
		return attachment, nil
	}
	f.storage.MoveFunc = func(context.Context, string, string, string) (attachmentout.ObjectInfo, error) {
		t.Error("Move() called for a failed attachment")
		return attachmentout.ObjectInfo{}, nil
	}
	var status domain.ReconciliationStatus
	f.outbox.RecordFailureFunc = func(
		_ context.Context, _ string, s domain.ReconciliationStatus, _ int, _ time.Time, _ time.Time,
	) error {
		status = s
		return nil
	}

	if err := f.uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if status != domain.ReconciliationStatusDead {
		t.Errorf("status = %q, want dead", status)
	}
	if f.metrics.deadAdded != 1 {
		t.Errorf("dead-letter metric = %d, want 1", f.metrics.deadAdded)
	}
}

func TestReconcileOutboxDeletesMovedObjectWhenDeletedMidProcessing(t *testing.T) {
	f := newReconcileFixture()
	var deletedKey string
	f.storage.DeleteFunc = func(_ context.Context, key string) error {
		deletedKey = key
		return nil
	}
	calls := 0
	f.attachments.UpdateStatusFunc = func(_ context.Context, _ string, status domain.AttachmentStatus) error {
		calls++
		if calls == 1 {
			return nil // PROCESSING succeeds
		}
		return ErrAttachmentNotFound // deleted before READY update
	}
	done := false
	f.outbox.MarkDoneFunc = func(context.Context, string, time.Time) error {
		done = true
		return nil
	}

	if err := f.uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if deletedKey != attachmentmocks.DefaultStorageKey {
		t.Errorf("deleted key = %q, want %q", deletedKey, attachmentmocks.DefaultStorageKey)
	}
	if !done {
		t.Error("outbox event was not marked done")
	}
}

func TestReconcileOutboxHandlesDeletingStatus(t *testing.T) {
	f := newReconcileFixture()
	att := pendingReconcileAttachment()
	if err := att.MarkProcessing(); err != nil {
		t.Fatalf("MarkProcessing: %v", err)
	}
	if err := att.MarkDeleting(); err != nil {
		t.Fatalf("MarkDeleting: %v", err)
	}
	f.attachments.GetByIDFunc = func(context.Context, string) (*domain.Attachment, error) {
		return att, nil
	}
	var deleted []string
	f.storage.DeleteFunc = func(_ context.Context, key string) error {
		deleted = append(deleted, key)
		return nil
	}
	var rowDeleted, done bool
	f.attachments.DeleteFunc = func(context.Context, string) error {
		rowDeleted = true
		return nil
	}
	f.outbox.MarkDoneFunc = func(context.Context, string, time.Time) error {
		done = true
		return nil
	}

	if err := f.uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(deleted) != 2 || !rowDeleted || !done {
		t.Errorf("cleanup incomplete: deleted=%v rowDeleted=%v done=%v", deleted, rowDeleted, done)
	}
}

func TestReconcileOutboxDeadLettersWhenBothObjectsMissing(t *testing.T) {
	f := newReconcileFixture()
	f.storage.MoveFunc = func(context.Context, string, string, string) (attachmentout.ObjectInfo, error) {
		return attachmentout.ObjectInfo{}, attachmentout.ErrObjectNotFound
	}
	var status domain.ReconciliationStatus
	f.outbox.RecordFailureFunc = func(
		_ context.Context, _ string, s domain.ReconciliationStatus, _ int, _ time.Time, _ time.Time,
	) error {
		status = s
		return nil
	}

	if err := f.uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if status != domain.ReconciliationStatusDead {
		t.Errorf("status = %q, want dead", status)
	}
	if len(f.statusWrites) < 2 || f.statusWrites[len(f.statusWrites)-1] != domain.AttachmentStatusFailed {
		t.Errorf("status writes = %v, want final failed", f.statusWrites)
	}
	if f.metrics.deadAdded != 1 {
		t.Errorf("dead-letter metric = %d, want 1", f.metrics.deadAdded)
	}
}

func TestReconcileOutboxStopsWhenNothingIsDue(t *testing.T) {
	f := newReconcileFixture()
	f.outbox.ClaimDueFunc = func(context.Context, time.Time) (*domain.ReconciliationEvent, error) {
		return nil, attachmentout.ErrNoDueEvents
	}
	f.storage.MoveFunc = func(context.Context, string, string, string) (attachmentout.ObjectInfo, error) {
		t.Error("Move() called with no due event")
		return attachmentout.ObjectInfo{}, nil
	}

	if err := f.uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(f.statusWrites) != 0 {
		t.Errorf("status writes = %v, want none", f.statusWrites)
	}
}

func TestReconcileOutboxPropagatesTransactionErrors(t *testing.T) {
	f := newReconcileFixture()
	wantErr := errors.New("database unavailable")
	f.outbox.ClaimDueFunc = func(context.Context, time.Time) (*domain.ReconciliationEvent, error) {
		return nil, wantErr
	}

	if err := f.uc.Execute(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want %v", err, wantErr)
	}
}

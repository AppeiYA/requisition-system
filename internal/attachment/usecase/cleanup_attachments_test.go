package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentmocks "github.com/AppeiYA/requisition-system/internal/attachment/mocks"
)

type stubMetricsRecorder struct {
	stagedPending int64
	outboxDepth   int64
	deadEvents    int64
	deletions     int64
	deadAdded     int64
	recorded      bool
}

func (m *stubMetricsRecorder) RecordQueueStats(_ context.Context, stagedPending, outboxDepth, deadEvents int64) {
	m.stagedPending = stagedPending
	m.outboxDepth = outboxDepth
	m.deadEvents = deadEvents
	m.recorded = true
}

func (m *stubMetricsRecorder) AddCleanupDeletions(_ context.Context, count int64) {
	m.deletions += count
}

func (m *stubMetricsRecorder) AddDeadLettered(_ context.Context, count int64) {
	m.deadAdded += count
}

func TestCleanupAttachmentsExpiresPendingAndRecoversStuckProcessing(t *testing.T) {
	now := time.Date(2026, 11, 27, 13, 0, 0, 0, time.UTC)
	staged := attachmentmocks.NewMockStagedUploadRepository()
	attachments := attachmentmocks.NewMockAttachmentRepository()
	outbox := attachmentmocks.NewMockReconciliationOutboxRepository()
	storage := attachmentmocks.NewMockFileStorage()
	tx := attachmentmocks.NewMockTransactionManager()
	metrics := &stubMetricsRecorder{}

	expiredUpload := attachmentmocks.ValidStagedUpload()
	staged.ExpirePendingFunc = func(_ context.Context, _ time.Time, _ int) ([]*domain.StagedUpload, error) {
		return []*domain.StagedUpload{expiredUpload}, nil
	}
	staged.CountPendingAllFunc = func(context.Context, time.Time) (int64, error) {
		return 3, nil
	}

	var deletedKeys []string
	storage.DeleteFunc = func(_ context.Context, key string) error {
		deletedKeys = append(deletedKeys, key)
		return nil
	}

	attachments.ResetStuckProcessingFunc = func(_ context.Context, stuckBefore time.Time, _ int) ([]string, error) {
		if !stuckBefore.Equal(now.Add(-15 * time.Minute)) {
			t.Errorf("stuckBefore = %v, want %v", stuckBefore, now.Add(-15*time.Minute))
		}
		return []string{"att-stuck-1"}, nil
	}

	var requeued []string
	outbox.RequeueByAttachmentIDsFunc = func(_ context.Context, ids []string, _ time.Time) (int64, error) {
		requeued = append(requeued, ids...)
		return int64(len(ids)), nil
	}
	outbox.CountDepthFunc = func(context.Context) (int64, int64, error) {
		return 4, 1, nil
	}

	uc := NewCleanupAttachmentsUsecase(
		staged, attachments, outbox, storage, tx, metrics, 50, 15*time.Minute,
	)
	uc.now = func() time.Time { return now }

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(deletedKeys) != 1 || deletedKeys[0] != expiredUpload.StoragePath() {
		t.Errorf("deleted keys = %v, want [%s]", deletedKeys, expiredUpload.StoragePath())
	}
	if len(requeued) != 1 || requeued[0] != "att-stuck-1" {
		t.Errorf("requeued = %v, want [att-stuck-1]", requeued)
	}
	if !metrics.recorded || metrics.stagedPending != 3 || metrics.outboxDepth != 4 || metrics.deadEvents != 1 || metrics.deletions != 1 {
		t.Errorf("unexpected metrics: %+v", metrics)
	}
}

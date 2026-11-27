package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentmocks "github.com/AppeiYA/requisition-system/internal/attachment/mocks"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
)

func claimFixture() (
	*ClaimStagedAttachmentsUsecase,
	*attachmentmocks.MockStagedUploadRepository,
	*attachmentmocks.MockAttachmentRepository,
	*attachmentmocks.MockReconciliationOutboxRepository,
	*[]string,
) {
	staged := attachmentmocks.NewMockStagedUploadRepository()
	attachments := attachmentmocks.NewMockAttachmentRepository()
	outbox := attachmentmocks.NewMockReconciliationOutboxRepository()
	claimed := &[]string{}
	staged.ClaimByIDsFunc = func(_ context.Context, ids []string, _ string, _ time.Time) ([]*domain.StagedUpload, error) {
		*claimed = append(*claimed, ids...)
		return []*domain.StagedUpload{attachmentmocks.ValidStagedUpload()}, nil
	}
	return NewClaimStagedAttachmentsUsecase(staged, attachments, outbox, attachmentmocks.NewMockIDGenerator()),
		staged, attachments, outbox, claimed
}

func TestClaimStagedAttachmentsCreatesPendingMetadataAndOutbox(t *testing.T) {
	uc, _, attachments, outbox, claimed := claimFixture()
	var saved *domain.Attachment
	var event *domain.ReconciliationEvent
	attachments.SaveFunc = func(_ context.Context, attachment *domain.Attachment) error {
		saved = attachment
		return nil
	}
	outbox.SaveFunc = func(_ context.Context, e *domain.ReconciliationEvent) error {
		event = e
		return nil
	}

	err := uc.Claim(context.Background(), []attachmentin.ClaimStagedAttachmentTarget{{
		StagedUploadID: attachmentmocks.DefaultStagedID,
		ActorID:        attachmentmocks.DefaultActorID,
		RefName:        "requisition_item",
		RefID:          "reqitem-5678",
	}})
	if err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	if saved == nil || saved.Status() != domain.AttachmentStatusPending {
		t.Fatalf("saved attachment status = %v, want pending", saved)
	}
	if saved.StagedUploadID() != attachmentmocks.DefaultStagedID {
		t.Errorf("staged provenance = %q", saved.StagedUploadID())
	}
	if want := "requisition_item/reqitem-5678/" + saved.ID(); saved.StorageKey() != want {
		t.Errorf("storage key = %q, want %q", saved.StorageKey(), want)
	}
	if event == nil || event.AttachmentID() != saved.ID() || event.StagedUploadID() != attachmentmocks.DefaultStagedID {
		t.Fatalf("outbox event does not reference the new attachment: %+v", event)
	}
	if event.FromKey() != attachmentmocks.DefaultStagingPath || event.ToKey() != saved.StorageKey() {
		t.Errorf("event keys = %q -> %q", event.FromKey(), event.ToKey())
	}
	if len(*claimed) != 1 || (*claimed)[0] != attachmentmocks.DefaultStagedID {
		t.Errorf("claimed IDs = %v", *claimed)
	}
}

func TestClaimStagedAttachmentsRejectsInvalidTargets(t *testing.T) {
	tests := []struct {
		name   string
		target attachmentin.ClaimStagedAttachmentTarget
	}{
		{
			name: "malformed token",
			target: attachmentin.ClaimStagedAttachmentTarget{
				StagedUploadID: "att-1234", ActorID: attachmentmocks.DefaultActorID,
				RefName: "requisition", RefID: attachmentmocks.DefaultRefID,
			},
		},
		{
			name: "missing actor",
			target: attachmentin.ClaimStagedAttachmentTarget{
				StagedUploadID: attachmentmocks.DefaultStagedID,
				RefName:        "requisition", RefID: attachmentmocks.DefaultRefID,
			},
		},
		{
			name: "wrong parent prefix",
			target: attachmentin.ClaimStagedAttachmentTarget{
				StagedUploadID: attachmentmocks.DefaultStagedID, ActorID: attachmentmocks.DefaultActorID,
				RefName: "requisition", RefID: "item-1234",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, _, _, _, claimed := claimFixture()
			err := uc.Claim(context.Background(), []attachmentin.ClaimStagedAttachmentTarget{tt.target})
			if !errors.Is(err, ErrStagedUploadUnavailable) && tt.name != "wrong parent prefix" {
				t.Fatalf("Claim() error = %v, want unavailable upload", err)
			}
			if err == nil {
				t.Fatal("Claim() unexpectedly succeeded")
			}
			if len(*claimed) != 0 {
				t.Errorf("claimed IDs = %v, want none", *claimed)
			}
		})
	}
}

func TestClaimStagedAttachmentsRejectsDuplicateTokens(t *testing.T) {
	uc, _, _, _, claimed := claimFixture()
	target := attachmentin.ClaimStagedAttachmentTarget{
		StagedUploadID: attachmentmocks.DefaultStagedID,
		ActorID:        attachmentmocks.DefaultActorID,
		RefName:        "requisition",
		RefID:          attachmentmocks.DefaultRefID,
	}
	err := uc.Claim(context.Background(), []attachmentin.ClaimStagedAttachmentTarget{target, target})
	if err != ErrStagedUploadUnavailable {
		t.Fatalf("Claim() error = %v, want %v", err, ErrStagedUploadUnavailable)
	}
	if len(*claimed) != 0 {
		t.Errorf("claimed IDs = %v, want none", *claimed)
	}
}

func TestClaimStagedAttachmentsRejectsUnavailableUploads(t *testing.T) {
	uc, staged, _, _, _ := claimFixture()
	staged.ClaimByIDsFunc = func(context.Context, []string, string, time.Time) ([]*domain.StagedUpload, error) {
		return nil, nil
	}
	err := uc.Claim(context.Background(), []attachmentin.ClaimStagedAttachmentTarget{{
		StagedUploadID: attachmentmocks.DefaultStagedID,
		ActorID:        attachmentmocks.DefaultActorID,
		RefName:        "requisition",
		RefID:          attachmentmocks.DefaultRefID,
	}})
	if err != ErrStagedUploadUnavailable {
		t.Fatalf("Claim() error = %v, want %v", err, ErrStagedUploadUnavailable)
	}
}

func TestClaimStagedAttachmentsConcurrentDoubleClaimOnlyOneWins(t *testing.T) {
	var mu sync.Mutex
	upload := attachmentmocks.ValidStagedUpload()
	alreadyClaimed := false

	staged := attachmentmocks.NewMockStagedUploadRepository()
	staged.ClaimByIDsFunc = func(_ context.Context, ids []string, actorID string, now time.Time) ([]*domain.StagedUpload, error) {
		mu.Lock()
		defer mu.Unlock()
		if alreadyClaimed || upload.CanClaim(actorID, now) != nil {
			return nil, nil
		}
		alreadyClaimed = true
		return []*domain.StagedUpload{upload}, nil
	}

	uc := NewClaimStagedAttachmentsUsecase(
		staged,
		attachmentmocks.NewMockAttachmentRepository(),
		attachmentmocks.NewMockReconciliationOutboxRepository(),
		attachmentmocks.NewMockIDGenerator(),
	)

	target := []attachmentin.ClaimStagedAttachmentTarget{{
		StagedUploadID: attachmentmocks.DefaultStagedID,
		ActorID:        attachmentmocks.DefaultActorID,
		RefName:        "requisition",
		RefID:          attachmentmocks.DefaultRefID,
	}}

	errs := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			errs <- uc.Claim(context.Background(), target)
		}()
	}
	wg.Wait()
	close(errs)

	var successes, failures int
	for err := range errs {
		if err == nil {
			successes++
		} else if errors.Is(err, ErrStagedUploadUnavailable) {
			failures++
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if successes != 1 || failures != 1 {
		t.Fatalf("successes=%d failures=%d, want 1 and 1", successes, failures)
	}
}

func TestClaimStagedAttachmentsNormalisesConflicts(t *testing.T) {
	uc, _, attachments, _, _ := claimFixture()
	attachments.SaveFunc = func(context.Context, *domain.Attachment) error {
		return attachmentout.ErrAttachmentExists
	}
	err := uc.Claim(context.Background(), []attachmentin.ClaimStagedAttachmentTarget{{
		StagedUploadID: attachmentmocks.DefaultStagedID,
		ActorID:        attachmentmocks.DefaultActorID,
		RefName:        "requisition",
		RefID:          attachmentmocks.DefaultRefID,
	}})
	if err != ErrStagedUploadUnavailable {
		t.Fatalf("Claim() error = %v, want %v", err, ErrStagedUploadUnavailable)
	}
}

func TestClaimStagedAttachmentsSkipsEmptyTargetList(t *testing.T) {
	uc, _, _, _, claimed := claimFixture()
	if err := uc.Claim(context.Background(), nil); err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	if len(*claimed) != 0 {
		t.Errorf("claimed IDs = %v, want none", *claimed)
	}
}

func TestClaimStagedAttachmentsEnforcesMaxPerRef(t *testing.T) {
	staged := attachmentmocks.NewMockStagedUploadRepository()
	uc := NewClaimStagedAttachmentsUsecase(
		staged,
		attachmentmocks.NewMockAttachmentRepository(),
		attachmentmocks.NewMockReconciliationOutboxRepository(),
		attachmentmocks.NewMockIDGenerator(),
		1,
	)
	err := uc.Claim(context.Background(), []attachmentin.ClaimStagedAttachmentTarget{
		{StagedUploadID: "upl-1", ActorID: attachmentmocks.DefaultActorID, RefName: "requisition", RefID: attachmentmocks.DefaultRefID},
		{StagedUploadID: "upl-2", ActorID: attachmentmocks.DefaultActorID, RefName: "requisition", RefID: attachmentmocks.DefaultRefID},
	})
	if err != ErrAttachmentLimitReached {
		t.Fatalf("Claim() error = %v, want ErrAttachmentLimitReached", err)
	}
}

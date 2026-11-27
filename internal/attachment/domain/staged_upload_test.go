package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func validStagedUpload(t *testing.T) *StagedUpload {
	t.Helper()
	name, err := NewAttachmentFileName("quote.pdf")
	if err != nil {
		t.Fatalf("NewAttachmentFileName() error = %v", err)
	}
	upload, err := NewStagedUpload(
		"upl-0123456789abcdef",
		"usr-0123456789abcdef",
		"staging/upl-0123456789abcdef",
		name,
		ContentTypePDF,
		42,
		strings.Repeat("a", ChecksumLength),
		time.Hour,
		time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("NewStagedUpload() error = %v", err)
	}
	return upload
}

func TestNewStagingUploadKey(t *testing.T) {
	key, err := NewStagingUploadKey("upl-0123456789abcdef")
	if err != nil {
		t.Fatalf("NewStagingUploadKey() error = %v", err)
	}
	if key != "staging/upl-0123456789abcdef" {
		t.Errorf("key = %q, want staging/upl-0123456789abcdef", key)
	}
	if _, err := NewStagingUploadKey("att-0123456789abcdef"); err == nil {
		t.Error("an attachment ID must not be accepted as a staged upload ID")
	}
}

func TestStagedUploadClaimGuards(t *testing.T) {
	upload := validStagedUpload(t)
	now := upload.CreatedAt().Add(time.Minute)

	if err := upload.CanClaim(upload.UploadedBy(), now); err != nil {
		t.Fatalf("CanClaim() error = %v", err)
	}
	if !errors.Is(upload.CanClaim("usr-other", now), ErrStagedUploadOwnerMismatch) {
		t.Error("a different actor was allowed to claim the upload")
	}
	if !errors.Is(upload.CanClaim(upload.UploadedBy(), upload.ExpiresAt()), ErrStagedUploadExpired) {
		t.Error("an expired upload was allowed to be claimed")
	}

	if err := upload.MarkClaimed(); err != nil {
		t.Fatalf("MarkClaimed() error = %v", err)
	}
	if upload.Status() != StagedUploadStatusClaimed {
		t.Errorf("status = %q, want claimed", upload.Status())
	}
	if !errors.Is(upload.CanClaim(upload.UploadedBy(), now), ErrStagedUploadNotPending) {
		t.Error("a claimed upload was claimable twice")
	}
}

func TestReconstituteStagedUploadRejectsBadRows(t *testing.T) {
	base := validStagedUpload(t)
	_, err := ReconstituteStagedUpload(
		base.ID(), base.UploadedBy(), base.StoragePath(), "quote.pdf",
		ContentTypePDF.String(), base.FileSize(), base.Checksum(), "ready",
		base.CreatedAt(), base.ExpiresAt(), nil,
	)
	if !errors.Is(err, ErrInvalidStagedUploadStatus) {
		t.Fatalf("error = %v, want ErrInvalidStagedUploadStatus", err)
	}
}

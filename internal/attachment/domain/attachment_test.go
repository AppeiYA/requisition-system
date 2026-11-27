package domain

import (
	"strings"
	"testing"
	"time"
)

func validAttachment(t *testing.T) *Attachment {
	t.Helper()
	ref, err := NewAttachmentRef(RefNameRequisition, "req-0123456789abcdef")
	if err != nil {
		t.Fatalf("NewAttachmentRef() error = %v", err)
	}
	name, err := NewAttachmentFileName("quote.pdf")
	if err != nil {
		t.Fatalf("NewAttachmentFileName() error = %v", err)
	}
	attachment, err := NewAttachment(
		"att-0123456789abcdef", ref, "requisition/req-0123456789abcdef/att-0123456789abcdef",
		name, ContentTypePDF, 42, strings.Repeat("a", ChecksumLength), "usr-0123456789abcdef",
	)
	if err != nil {
		t.Fatalf("NewAttachment() error = %v", err)
	}
	return attachment
}

func TestNewAttachmentRefUsesExplicitSeparator(t *testing.T) {
	if _, err := NewAttachmentRef(RefNameRequisition, "reqitem-0123456789abcdef"); err == nil {
		t.Error("a requisition ref must not accept a line ID")
	}
	if _, err := NewAttachmentRef(RefNameRequisitionItem, "req-0123456789abcdef"); err == nil {
		t.Error("a line ref must not accept a requisition ID")
	}
	if _, err := NewAttachmentRef(RefNameRequisition, "req"); err == nil {
		t.Error("a bare prefix must be rejected")
	}

	ref, err := NewAttachmentRef(RefNameRequisitionItem, "reqitem-0123456789abcdef")
	if err != nil {
		t.Fatalf("NewAttachmentRef() error = %v", err)
	}
	if ref.RefName() != RefNameRequisitionItem {
		t.Errorf("RefName() = %q, want %q", ref.RefName(), RefNameRequisitionItem)
	}
}

func TestNewAttachmentKey(t *testing.T) {
	attachment := validAttachment(t)
	key, err := NewAttachmentKey(attachment.Ref(), attachment.ID())
	if err != nil {
		t.Fatalf("NewAttachmentKey() error = %v", err)
	}
	if want := "requisition/req-0123456789abcdef/att-0123456789abcdef"; key != want {
		t.Errorf("key = %q, want %q", key, want)
	}
	if strings.Contains(key, "..") {
		t.Error("key contains a traversal segment")
	}
}

func TestNewAttachmentValidatesFields(t *testing.T) {
	base := validAttachment(t)
	ref := base.Ref()
	name := base.FileName()

	tests := []struct {
		name     string
		id       string
		size     int64
		checksum string
		actor    string
	}{
		{name: "missing ID separator", id: "att", size: 1, checksum: strings.Repeat("a", 64), actor: "usr-1"},
		{name: "non-positive size", id: "att-1", size: 0, checksum: strings.Repeat("a", 64), actor: "usr-1"},
		{name: "bad checksum", id: "att-1", size: 1, checksum: "abc", actor: "usr-1"},
		{name: "bare user prefix", id: "att-1", size: 1, checksum: strings.Repeat("a", 64), actor: "usr"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewAttachment(tt.id, ref, "key", name, ContentTypePDF, tt.size, tt.checksum, tt.actor); err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestReconstituteAttachmentUsesExplicitRefName(t *testing.T) {
	got, err := ReconstituteAttachment(
		"att-0123456789abcdef", "requisition_item", "reqitem-0123456789abcdef",
		"requisition_item/reqitem-0123456789abcdef/att-0123456789abcdef",
		"photo.png", "image/png", 12, strings.Repeat("b", ChecksumLength),
		"usr-0123456789abcdef", time.Unix(0, 0).UTC(),
		AttachmentStatusReady.String(), "",
	)
	if err != nil {
		t.Fatalf("ReconstituteAttachment() error = %v", err)
	}
	if got.Ref().RefName() != RefNameRequisitionItem {
		t.Errorf("RefName() = %q, want %q", got.Ref().RefName(), RefNameRequisitionItem)
	}
	if !got.CreatedAt().Equal(time.Unix(0, 0).UTC()) {
		t.Errorf("CreatedAt() = %v", got.CreatedAt())
	}
}

func TestAttachmentLifecycleGate(t *testing.T) {
	attachment := validAttachment(t)
	if attachment.Status() != AttachmentStatusReady {
		t.Fatalf("direct upload status = %q, want ready", attachment.Status())
	}
	if err := attachment.CanDownload(); err != nil {
		t.Fatalf("ready attachment cannot be downloaded: %v", err)
	}

	ref, _ := NewAttachmentRef(RefNameRequisitionItem, "reqitem-0123456789abcdef")
	name, _ := NewAttachmentFileName("quote.pdf")
	pending, err := NewPendingAttachment(
		"att-fedcba9876543210", ref, "upl-0123456789abcdef",
		"requisition_item/reqitem-0123456789abcdef/att-fedcba9876543210",
		name, ContentTypePDF, 42, strings.Repeat("c", ChecksumLength),
		"usr-0123456789abcdef",
	)
	if err != nil {
		t.Fatalf("NewPendingAttachment() error = %v", err)
	}
	if err := pending.CanDownload(); err != ErrAttachmentNotReady {
		t.Errorf("pending CanDownload() error = %v, want %v", err, ErrAttachmentNotReady)
	}
	if err := pending.MarkReady(); err == nil {
		t.Error("pending attachment must not jump straight to ready")
	}
	if err := pending.MarkProcessing(); err != nil {
		t.Fatalf("MarkProcessing() error = %v", err)
	}
	if err := pending.MarkReady(); err != nil {
		t.Fatalf("MarkReady() error = %v", err)
	}
	if err := pending.CanDownload(); err != nil {
		t.Fatalf("reconciled attachment cannot be downloaded: %v", err)
	}
}

func TestIsInlineSafe(t *testing.T) {
	for _, ct := range []ContentType{ContentTypePDF, ContentTypePNG, ContentTypeJPEG, ContentTypeGIF} {
		if !ct.IsInlineSafe() {
			t.Errorf("%s should be inline safe", ct)
		}
	}
	for _, ct := range []ContentType{ContentTypeDOCX, ContentTypeXLSX, ContentTypeCSV} {
		if ct.IsInlineSafe() {
			t.Errorf("%s must not be inline safe", ct)
		}
	}
}

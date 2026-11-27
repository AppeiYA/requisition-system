package domain

import (
	"encoding/hex"
	"strings"
	"time"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var AttachmentIDPrefix = "att"

var (
	ErrInvalidAttachmentID      = customerrors.NewBadRequestError("invalid attachment ID")
	ErrInvalidAttachmentRef     = customerrors.NewBadRequestError("invalid attachment ref")
	ErrInvalidStorageKey        = customerrors.NewBadRequestError("invalid attachment storage key")
	ErrInvalidChecksum          = customerrors.NewBadRequestError("invalid attachment checksum")
	ErrInvalidAttachmentContent = customerrors.NewBadRequestError("invalid attachment content type")
	ErrInvalidAttachmentSize    = customerrors.NewBadRequestError("attachment file size must be positive")
	ErrInvalidUploadActor       = customerrors.NewBadRequestError("invalid upload actor")
	ErrInvalidAttachmentStatus  = customerrors.NewBadRequestError("invalid attachment status")
	ErrInvalidStagedProvenance  = customerrors.NewBadRequestError("invalid staged upload provenance")
	ErrAttachmentNotReady       = customerrors.NewConflictError("attachment is not ready for download")
	ErrAttachmentStatusConflict = customerrors.NewConflictError("attachment status transition is not allowed")
)

const ChecksumLength = 64

// AttachmentStatus tracks reconciliation. Directly uploaded attachments are
// READY on creation; ones claimed at requisition creation start at PENDING and
// only become downloadable once the outbox worker moves their bytes.
type AttachmentStatus string

const (
	AttachmentStatusPending    AttachmentStatus = "pending"
	AttachmentStatusProcessing AttachmentStatus = "processing"
	AttachmentStatusReady      AttachmentStatus = "ready"
	AttachmentStatusFailed     AttachmentStatus = "failed"
	AttachmentStatusDeleting   AttachmentStatus = "deleting"
)

func (s AttachmentStatus) IsValid() bool {
	switch s {
	case AttachmentStatusPending, AttachmentStatusProcessing,
		AttachmentStatusReady, AttachmentStatusFailed, AttachmentStatusDeleting:
		return true
	default:
		return false
	}
}

func (s AttachmentStatus) String() string { return string(s) }

type Attachment struct {
	id             string
	ref            *AttachmentRef
	storageKey     string
	fileName       *AttachmentFileName
	contentType    ContentType
	fileSize       int64
	checksum       string
	uploadedBy     string
	createdAt      time.Time
	status         AttachmentStatus
	stagedUploadID string
}

func NewAttachment(
	id string,
	ref *AttachmentRef,
	storageKey string,
	fileName *AttachmentFileName,
	contentType ContentType,
	fileSize int64,
	checksum string,
	uploadedBy string,
) (*Attachment, error) {
	if !isValidAttachmentID(id) {
		return nil, ErrInvalidAttachmentID
	}
	if ref == nil {
		return nil, ErrInvalidAttachmentRef
	}
	if storageKey == "" {
		return nil, ErrInvalidStorageKey
	}
	if fileName == nil {
		return nil, ErrInvalidAttachmentFileName
	}
	if !contentType.IsValid() {
		return nil, ErrInvalidAttachmentContent
	}
	if fileSize <= 0 {
		return nil, ErrInvalidAttachmentSize
	}
	if !isValidChecksum(checksum) {
		return nil, ErrInvalidChecksum
	}
	if !strings.HasPrefix(uploadedBy, "usr-") {
		return nil, ErrInvalidUploadActor
	}

	return &Attachment{
		id:          id,
		ref:         ref,
		storageKey:  storageKey,
		fileName:    fileName,
		contentType: contentType,
		fileSize:    fileSize,
		checksum:    checksum,
		uploadedBy:  uploadedBy,
		createdAt:   time.Now(),
		status:      AttachmentStatusReady,
	}, nil
}

// NewPendingAttachment creates the metadata half of a creation-time
// attachment. Its bytes still live under staging/<stagedUploadID> until the
// outbox worker claims them, so the row is deliberately not READY yet.
func NewPendingAttachment(
	id string,
	ref *AttachmentRef,
	stagedUploadID string,
	storageKey string,
	fileName *AttachmentFileName,
	contentType ContentType,
	fileSize int64,
	checksum string,
	uploadedBy string,
) (*Attachment, error) {
	if !IsValidStagedUploadID(stagedUploadID) {
		return nil, ErrInvalidStagedProvenance
	}
	attachment, err := NewAttachment(
		id, ref, storageKey, fileName, contentType, fileSize, checksum, uploadedBy,
	)
	if err != nil {
		return nil, err
	}
	attachment.status = AttachmentStatusPending
	attachment.stagedUploadID = stagedUploadID
	return attachment, nil
}

// ReconstituteAttachment rebuilds an attachment from persisted columns. The
// ref name is stored explicitly, so parent resolution never guesses a prefix.
func ReconstituteAttachment(
	id string,
	refName string,
	refID string,
	storageKey string,
	fileName string,
	contentType string,
	fileSize int64,
	checksum string,
	uploadedBy string,
	createdAt time.Time,
	status string,
	stagedUploadID string,
) (*Attachment, error) {
	ref, err := NewAttachmentRef(AttachmentRefName(refName), refID)
	if err != nil {
		return nil, err
	}
	name, err := NewAttachmentFileName(fileName)
	if err != nil {
		return nil, err
	}
	ct := ContentType(contentType)
	if !ct.IsValid() {
		return nil, ErrInvalidAttachmentContent
	}
	if !isValidAttachmentID(id) {
		return nil, ErrInvalidAttachmentID
	}
	if storageKey == "" {
		return nil, ErrInvalidStorageKey
	}
	if fileSize <= 0 {
		return nil, ErrInvalidAttachmentSize
	}
	if !isValidChecksum(checksum) {
		return nil, ErrInvalidChecksum
	}
	if !strings.HasPrefix(uploadedBy, "usr-") {
		return nil, ErrInvalidUploadActor
	}
	attachmentStatus := AttachmentStatus(status)
	if !attachmentStatus.IsValid() {
		return nil, ErrInvalidAttachmentStatus
	}
	if stagedUploadID != "" && !IsValidStagedUploadID(stagedUploadID) {
		return nil, ErrInvalidStagedProvenance
	}
	if attachmentStatus != AttachmentStatusReady && stagedUploadID == "" {
		return nil, ErrInvalidStagedProvenance
	}

	return &Attachment{
		id:             id,
		ref:            ref,
		storageKey:     storageKey,
		fileName:       name,
		contentType:    ct,
		fileSize:       fileSize,
		checksum:       checksum,
		uploadedBy:     uploadedBy,
		createdAt:      createdAt,
		status:         attachmentStatus,
		stagedUploadID: stagedUploadID,
	}, nil
}

func isValidAttachmentID(id string) bool {
	return strings.HasPrefix(id, AttachmentIDPrefix+"-") && id != AttachmentIDPrefix+"-"
}

func isValidChecksum(checksum string) bool {
	if len(checksum) != ChecksumLength {
		return false
	}
	_, err := hex.DecodeString(checksum)
	return err == nil && checksum == strings.ToLower(checksum)
}

func (a *Attachment) ID() string                    { return a.id }
func (a *Attachment) Ref() *AttachmentRef           { return a.ref }
func (a *Attachment) StorageKey() string            { return a.storageKey }
func (a *Attachment) FileName() *AttachmentFileName { return a.fileName }
func (a *Attachment) ContentType() ContentType      { return a.contentType }
func (a *Attachment) FileSize() int64               { return a.fileSize }
func (a *Attachment) Checksum() string              { return a.checksum }
func (a *Attachment) UploadedBy() string            { return a.uploadedBy }
func (a *Attachment) CreatedAt() time.Time          { return a.createdAt }
func (a *Attachment) Status() AttachmentStatus      { return a.status }
func (a *Attachment) StagedUploadID() string        { return a.stagedUploadID }

// CanDownload is the single gate between a metadata row and its bytes. Pending,
// processing and failed rows may be listed, but never served.
func (a *Attachment) CanDownload() error {
	if a == nil || a.status != AttachmentStatusReady {
		return ErrAttachmentNotReady
	}
	return nil
}

// MarkProcessing and its siblings are used by the reconciliation worker. They
// are kept on the aggregate so no use case can write an arbitrary status.
func (a *Attachment) MarkProcessing() error {
	return a.transition(AttachmentStatusPending, AttachmentStatusProcessing)
}

func (a *Attachment) MarkReady() error {
	return a.transition(AttachmentStatusProcessing, AttachmentStatusReady)
}

func (a *Attachment) MarkFailed() error {
	if a == nil || (a.status != AttachmentStatusPending && a.status != AttachmentStatusProcessing) {
		return ErrAttachmentStatusConflict
	}
	a.status = AttachmentStatusFailed
	return nil
}

func (a *Attachment) MarkDeleting() error {
	if a == nil || (a.status != AttachmentStatusPending && a.status != AttachmentStatusProcessing) {
		return ErrAttachmentStatusConflict
	}
	a.status = AttachmentStatusDeleting
	return nil
}

func (a *Attachment) transition(from, to AttachmentStatus) error {
	if a == nil || a.status != from {
		return ErrAttachmentStatusConflict
	}
	a.status = to
	return nil
}

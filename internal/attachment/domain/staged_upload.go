package domain

import (
	"strings"
	"time"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

// StagedUploadIDPrefix is the application-generated prefix for a file that
// exists before its parent requisition or requisition line does.
var StagedUploadIDPrefix = "upl"

const StagingPathPrefix = "staging"

var (
	ErrInvalidStagedUploadID       = customerrors.NewBadRequestError("invalid staged upload ID")
	ErrInvalidStagedUploadPath     = customerrors.NewBadRequestError("invalid staged upload storage path")
	ErrInvalidStagedUploadStatus   = customerrors.NewBadRequestError("invalid staged upload status")
	ErrInvalidStagedUploadExpiry   = customerrors.NewBadRequestError("staged upload expiry must be in the future")
	ErrStagedUploadNotPending      = customerrors.NewConflictError("staged upload is no longer pending")
	ErrStagedUploadOwnerMismatch   = customerrors.NewForbiddenError("staged upload belongs to another user")
	ErrStagedUploadExpired         = customerrors.NewConflictError("staged upload has expired")
	ErrInvalidStagedUploadChecksum = customerrors.NewBadRequestError("invalid staged upload checksum")
	ErrInvalidStagedUploadRow      = customerrors.NewInternalServerError("invalid staged upload row")
)

// StagedUploadStatus is the lifecycle of a parentless upload. Claimed rows are
// kept until reconciliation finishes; expired rows are collected by cleanup.
type StagedUploadStatus string

const (
	StagedUploadStatusPending StagedUploadStatus = "pending"
	StagedUploadStatusClaimed StagedUploadStatus = "claimed"
	StagedUploadStatusExpired StagedUploadStatus = "expired"
	StagedUploadStatusDeleted StagedUploadStatus = "deleted"
)

func (s StagedUploadStatus) IsValid() bool {
	switch s {
	case StagedUploadStatusPending, StagedUploadStatusClaimed,
		StagedUploadStatusExpired, StagedUploadStatusDeleted:
		return true
	default:
		return false
	}
}

func (s StagedUploadStatus) String() string { return string(s) }

// StagedUpload is a file whose bytes are stored but whose permanent parent has
// not been established yet. StoragePath is always staging/<upload-id>; no
// client string contributes to it.
type StagedUpload struct {
	id               string
	uploadedBy       string
	storagePath      string
	originalFileName *AttachmentFileName
	contentType      ContentType
	fileSize         int64
	checksum         string
	status           StagedUploadStatus
	createdAt        time.Time
	expiresAt        time.Time
	claimedAt        *time.Time
}

func NewStagedUpload(
	id string,
	uploadedBy string,
	storagePath string,
	fileName *AttachmentFileName,
	contentType ContentType,
	fileSize int64,
	checksum string,
	ttl time.Duration,
	now time.Time,
) (*StagedUpload, error) {
	if !IsValidStagedUploadID(id) {
		return nil, ErrInvalidStagedUploadID
	}
	if !strings.HasPrefix(uploadedBy, "usr-") {
		return nil, ErrInvalidUploadActor
	}
	if want, err := NewStagingUploadKey(id); err != nil || storagePath != want {
		return nil, ErrInvalidStagedUploadPath
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
		return nil, ErrInvalidStagedUploadChecksum
	}
	if ttl <= 0 || !now.Add(ttl).After(now) {
		return nil, ErrInvalidStagedUploadExpiry
	}

	return &StagedUpload{
		id:               id,
		uploadedBy:       uploadedBy,
		storagePath:      storagePath,
		originalFileName: fileName,
		contentType:      contentType,
		fileSize:         fileSize,
		checksum:         checksum,
		status:           StagedUploadStatusPending,
		createdAt:        now.UTC(),
		expiresAt:        now.UTC().Add(ttl),
	}, nil
}

// ReconstituteStagedUpload rebuilds a persisted row. Statuses are explicit in
// the database; nothing is inferred from ID prefixes or file names.
func ReconstituteStagedUpload(
	id string,
	uploadedBy string,
	storagePath string,
	fileName string,
	contentType string,
	fileSize int64,
	checksum string,
	status string,
	createdAt time.Time,
	expiresAt time.Time,
	claimedAt *time.Time,
) (*StagedUpload, error) {
	name, err := NewAttachmentFileName(fileName)
	if err != nil {
		return nil, err
	}
	ct := ContentType(contentType)
	if !ct.IsValid() {
		return nil, ErrInvalidAttachmentContent
	}
	stagedStatus := StagedUploadStatus(status)
	if !stagedStatus.IsValid() {
		return nil, ErrInvalidStagedUploadStatus
	}
	if !IsValidStagedUploadID(id) || !strings.HasPrefix(uploadedBy, "usr-") {
		return nil, ErrInvalidStagedUploadID
	}
	if want, err := NewStagingUploadKey(id); err != nil || storagePath != want {
		return nil, ErrInvalidStagedUploadPath
	}
	if fileSize <= 0 || !isValidChecksum(checksum) {
		return nil, ErrInvalidStagedUploadRow
	}
	if expiresAt.IsZero() {
		return nil, ErrInvalidStagedUploadExpiry
	}

	var claimedCopy *time.Time
	if claimedAt != nil {
		t := claimedAt.UTC()
		claimedCopy = &t
	}

	return &StagedUpload{
		id:               id,
		uploadedBy:       uploadedBy,
		storagePath:      storagePath,
		originalFileName: name,
		contentType:      ct,
		fileSize:         fileSize,
		checksum:         checksum,
		status:           stagedStatus,
		createdAt:        createdAt,
		expiresAt:        expiresAt,
		claimedAt:        claimedCopy,
	}, nil
}

// CanClaim enforces the three guards a requisition creation must satisfy
// before it is allowed to turn this file into a permanent attachment.
func (u *StagedUpload) CanClaim(actorID string, now time.Time) error {
	if u == nil {
		return ErrInvalidStagedUploadID
	}
	if u.status != StagedUploadStatusPending {
		return ErrStagedUploadNotPending
	}
	if u.uploadedBy != actorID {
		return ErrStagedUploadOwnerMismatch
	}
	if !u.expiresAt.After(now) {
		return ErrStagedUploadExpired
	}
	return nil
}

// MarkClaimed is called inside the requisition transaction together with the
// attachment metadata write. It is deliberately not exported as a standalone
// use case: claiming without a parent is never valid.
func (u *StagedUpload) MarkClaimed() error {
	if u == nil || u.status != StagedUploadStatusPending {
		return ErrStagedUploadNotPending
	}
	u.status = StagedUploadStatusClaimed
	return nil
}

func (u *StagedUpload) ID() string                            { return u.id }
func (u *StagedUpload) UploadedBy() string                    { return u.uploadedBy }
func (u *StagedUpload) StoragePath() string                   { return u.storagePath }
func (u *StagedUpload) OriginalFileName() *AttachmentFileName { return u.originalFileName }
func (u *StagedUpload) ContentType() ContentType              { return u.contentType }
func (u *StagedUpload) FileSize() int64                       { return u.fileSize }
func (u *StagedUpload) Checksum() string                      { return u.checksum }
func (u *StagedUpload) Status() StagedUploadStatus            { return u.status }
func (u *StagedUpload) CreatedAt() time.Time                  { return u.createdAt }
func (u *StagedUpload) ExpiresAt() time.Time                  { return u.expiresAt }
func (u *StagedUpload) ClaimedAt() *time.Time {
	if u.claimedAt == nil {
		return nil
	}
	t := *u.claimedAt
	return &t
}

// NewStagingUploadKey derives the only key shape staging may use. It is also
// used by the storage adapter to validate keys before touching the filesystem.
func NewStagingUploadKey(id string) (string, error) {
	if !IsValidStagedUploadID(id) {
		return "", ErrInvalidStagedUploadID
	}
	return StagingPathPrefix + "/" + id, nil
}

func IsValidStagedUploadID(id string) bool {
	return strings.HasPrefix(id, StagedUploadIDPrefix+"-") && id != StagedUploadIDPrefix+"-"
}

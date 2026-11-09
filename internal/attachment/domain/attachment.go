package domain

import (
	"strings"
	"time"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var AttachmentIDPrefix = "att"

var (
	ErrInvalidAttachmentID = customerrors.NewBadRequestError("invalid attachment ID")
	ErrInvalidUploadActor  = customerrors.NewBadRequestError("invalid upload actor")
)

type Attachment struct {
	id string
	ref *AttachmentRef
	fileName *AttachmentFileName
	fileUrl *AttachmentFileUrl
	contentType ContentType
	fileSize int64
	uploadedBy string
	createdAt time.Time
}

func NewAttachment(id string, ref *AttachmentRef, fileName *AttachmentFileName, fileUrl *AttachmentFileUrl, contentType ContentType, fileSize int64, uploadedBy string) (*Attachment, error) {
	if id == "" || !strings.HasPrefix(id, AttachmentIDPrefix) {
        return nil, ErrInvalidAttachmentID
	}

	if uploadedBy == "" || !strings.HasPrefix(uploadedBy, "usr"){
		return nil, ErrInvalidUploadActor
	}

	createdAt := time.Now()

	return &Attachment{
		id: id,
		ref: ref,
		fileName: fileName,
		fileUrl: fileUrl,
		contentType: contentType,
		fileSize: fileSize,
		uploadedBy: uploadedBy,
		createdAt: createdAt,
	}, nil
}

func ReconstituteAttachment(
	id string,
	refID string,
	filename string,
	fileurl string,
	contentType string,
	fileSize int64,
	uploadedBy string,
	createdAt time.Time,
) (*Attachment, error) {
	ref, err := ReconstructAttachmentRef(refID)
	if err != nil {
		return nil, err
	}

	fileName, err := NewAttachmentFileName(filename)
	if err != nil {
		return nil, err
	}

	fileUrl, err := NewAttachmentFileUrl(fileurl)
	if err != nil {
		return nil, err
	}

	contentTypeValue := ContentType(contentType)
	if !contentTypeValue.IsValid() {
		return nil, customerrors.NewBadRequestError("invalid content type")
	}

	return &Attachment{
		id: id,
		ref: ref,
		fileName: fileName,
		fileUrl: fileUrl,
		contentType: contentTypeValue,
		fileSize: fileSize,
		uploadedBy: uploadedBy,
		createdAt: createdAt,
	}, nil
}

func (a *Attachment) ID() string {
	return a.id
}

func (a *Attachment) Ref() *AttachmentRef {
	return a.ref
}

func (a *Attachment) FileName() *AttachmentFileName {
	return a.fileName
}

func (a *Attachment) FileUrl() *AttachmentFileUrl {
	return a.fileUrl
}

func (a *Attachment) ContentType() ContentType {
	return a.contentType
}

func (a *Attachment) FileSize() int64 {
	return a.fileSize
}

func (a *Attachment) UploadedBy() string {
	return a.uploadedBy
}

func (a *Attachment) CreatedAt() time.Time {
	return a.createdAt
}
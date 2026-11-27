package models

import (
	"database/sql"
	"time"

	attachmentdomain "github.com/AppeiYA/requisition-system/internal/attachment/domain"
)

type AttachmentModel struct {
	ID             string         `db:"id"`
	RefName        string         `db:"ref_name"`
	RefID          string         `db:"ref_id"`
	StorageKey     string         `db:"storage_key"`
	FileName       string         `db:"file_name"`
	ContentType    string         `db:"content_type"`
	FileSize       int64          `db:"file_size"`
	Checksum       string         `db:"checksum"`
	UploadedBy     string         `db:"uploaded_by"`
	CreatedAt      time.Time      `db:"created_at"`
	Status         string         `db:"status"`
	StagedUploadID sql.NullString `db:"staged_upload_id"`
}

func AttachmentFromDomain(attachment *attachmentdomain.Attachment) *AttachmentModel {
	ref := attachment.Ref()
	stagedID := sql.NullString{}
	if attachment.StagedUploadID() != "" {
		stagedID = sql.NullString{String: attachment.StagedUploadID(), Valid: true}
	}
	return &AttachmentModel{
		ID:             attachment.ID(),
		RefName:        ref.RefName().String(),
		RefID:          ref.ID(),
		StorageKey:     attachment.StorageKey(),
		FileName:       attachment.FileName().String(),
		ContentType:    attachment.ContentType().String(),
		FileSize:       attachment.FileSize(),
		Checksum:       attachment.Checksum(),
		UploadedBy:     attachment.UploadedBy(),
		CreatedAt:      attachment.CreatedAt(),
		Status:         attachment.Status().String(),
		StagedUploadID: stagedID,
	}
}

func (m AttachmentModel) ToDomain() (*attachmentdomain.Attachment, error) {
	stagedUploadID := ""
	if m.StagedUploadID.Valid {
		stagedUploadID = m.StagedUploadID.String
	}
	return attachmentdomain.ReconstituteAttachment(
		m.ID,
		m.RefName,
		m.RefID,
		m.StorageKey,
		m.FileName,
		m.ContentType,
		m.FileSize,
		m.Checksum,
		m.UploadedBy,
		m.CreatedAt,
		m.Status,
		stagedUploadID,
	)
}

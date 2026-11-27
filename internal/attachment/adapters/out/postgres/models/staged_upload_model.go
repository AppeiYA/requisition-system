package models

import (
	"database/sql"
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
)

type StagedUploadModel struct {
	ID               string       `db:"id"`
	UploadedBy       string       `db:"uploaded_by"`
	StoragePath      string       `db:"storage_path"`
	OriginalFileName string       `db:"original_filename"`
	ContentType      string       `db:"content_type"`
	FileSize         int64        `db:"file_size"`
	Checksum         string       `db:"checksum"`
	Status           string       `db:"status"`
	CreatedAt        time.Time    `db:"created_at"`
	ExpiresAt        time.Time    `db:"expires_at"`
	ClaimedAt        sql.NullTime `db:"claimed_at"`
}

func StagedUploadFromDomain(upload *domain.StagedUpload) StagedUploadModel {
	var claimedAt sql.NullTime
	if c := upload.ClaimedAt(); c != nil {
		claimedAt = sql.NullTime{Time: *c, Valid: true}
	}
	return StagedUploadModel{
		ID:               upload.ID(),
		UploadedBy:       upload.UploadedBy(),
		StoragePath:      upload.StoragePath(),
		OriginalFileName: upload.OriginalFileName().String(),
		ContentType:      upload.ContentType().String(),
		FileSize:         upload.FileSize(),
		Checksum:         upload.Checksum(),
		Status:           upload.Status().String(),
		CreatedAt:        upload.CreatedAt(),
		ExpiresAt:        upload.ExpiresAt(),
		ClaimedAt:        claimedAt,
	}
}

func (m StagedUploadModel) ToDomain() (*domain.StagedUpload, error) {
	var claimedAt *time.Time
	if m.ClaimedAt.Valid {
		t := m.ClaimedAt.Time
		claimedAt = &t
	}
	return domain.ReconstituteStagedUpload(
		m.ID,
		m.UploadedBy,
		m.StoragePath,
		m.OriginalFileName,
		m.ContentType,
		m.FileSize,
		m.Checksum,
		m.Status,
		m.CreatedAt,
		m.ExpiresAt,
		claimedAt,
	)
}

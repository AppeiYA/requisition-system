package models

import (
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
)

type ReconciliationOutboxModel struct {
	ID             string    `db:"id"`
	AttachmentID   string    `db:"attachment_id"`
	StagedUploadID string    `db:"staged_upload_id"`
	FromKey        string    `db:"from_key"`
	ToKey          string    `db:"to_key"`
	Checksum       string    `db:"checksum"`
	Status         string    `db:"status"`
	Attempts       int       `db:"attempts"`
	NextAttemptAt  time.Time `db:"next_attempt_at"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

func ReconciliationOutboxFromDomain(event *domain.ReconciliationEvent) ReconciliationOutboxModel {
	return ReconciliationOutboxModel{
		ID:             event.ID(),
		AttachmentID:   event.AttachmentID(),
		StagedUploadID: event.StagedUploadID(),
		FromKey:        event.FromKey(),
		ToKey:          event.ToKey(),
		Checksum:       event.Checksum(),
		Status:         event.Status().String(),
		Attempts:       event.Attempts(),
		NextAttemptAt:  event.NextAttemptAt(),
		CreatedAt:      event.CreatedAt(),
		UpdatedAt:      event.UpdatedAt(),
	}
}

func (m ReconciliationOutboxModel) ToDomain() (*domain.ReconciliationEvent, error) {
	return domain.ReconstituteReconciliationEvent(
		m.ID,
		m.AttachmentID,
		m.StagedUploadID,
		m.FromKey,
		m.ToKey,
		m.Checksum,
		m.Status,
		m.Attempts,
		m.NextAttemptAt,
		m.CreatedAt,
		m.UpdatedAt,
	)
}

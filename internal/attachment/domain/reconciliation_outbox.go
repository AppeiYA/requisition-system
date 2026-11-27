package domain

import (
	"math/rand/v2"
	"time"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

// ReconciliationEventIDPrefix identifies an attachment reconciliation outbox
// row. The outbox is written inside the requisition transaction and processed
// asynchronously afterwards.
var ReconciliationEventIDPrefix = "aevt"

var (
	ErrInvalidReconciliationID    = customerrors.NewBadRequestError("invalid reconciliation event ID")
	ErrInvalidReconciliationEvent = customerrors.NewBadRequestError("invalid reconciliation event")
	ErrInvalidReconciliationState = customerrors.NewBadRequestError("invalid reconciliation event status")
)

// ReconciliationStatus follows the same lifecycle as the attachment it moves,
// with done marking a fully reconciled event and dead a permanently failed one.
type ReconciliationStatus string

const (
	ReconciliationStatusPending    ReconciliationStatus = "pending"
	ReconciliationStatusProcessing ReconciliationStatus = "processing"
	ReconciliationStatusDone       ReconciliationStatus = "done"
	ReconciliationStatusFailed     ReconciliationStatus = "failed"
	ReconciliationStatusDead       ReconciliationStatus = "dead"
)

func (s ReconciliationStatus) IsValid() bool {
	switch s {
	case ReconciliationStatusPending, ReconciliationStatusProcessing,
		ReconciliationStatusDone, ReconciliationStatusFailed, ReconciliationStatusDead:
		return true
	default:
		return false
	}
}

func (s ReconciliationStatus) String() string { return string(s) }

// ReconciliationEvent is the durable instruction to move one staged object to
// its permanent key. It contains only validated storage keys and IDs; the
// worker never needs to trust a client-supplied path.
type ReconciliationEvent struct {
	id             string
	attachmentID   string
	stagedUploadID string
	fromKey        string
	toKey          string
	checksum       string
	status         ReconciliationStatus
	attempts       int
	nextAttemptAt  time.Time
	createdAt      time.Time
	updatedAt      time.Time
}

func NewReconciliationEvent(
	id string,
	attachmentID string,
	stagedUploadID string,
	fromKey string,
	toKey string,
	checksum string,
	now time.Time,
) (*ReconciliationEvent, error) {
	if !isValidReconciliationID(id) {
		return nil, ErrInvalidReconciliationID
	}
	if !isValidAttachmentID(attachmentID) || !IsValidStagedUploadID(stagedUploadID) {
		return nil, ErrInvalidReconciliationEvent
	}
	if fromKey == "" || toKey == "" || !isValidChecksum(checksum) || now.IsZero() {
		return nil, ErrInvalidReconciliationEvent
	}

	return &ReconciliationEvent{
		id:             id,
		attachmentID:   attachmentID,
		stagedUploadID: stagedUploadID,
		fromKey:        fromKey,
		toKey:          toKey,
		checksum:       checksum,
		status:         ReconciliationStatusPending,
		attempts:       0,
		nextAttemptAt:  now.UTC(),
		createdAt:      now.UTC(),
		updatedAt:      now.UTC(),
	}, nil
}

// ReconstituteReconciliationEvent rebuilds an event row for the worker.
func ReconstituteReconciliationEvent(
	id string,
	attachmentID string,
	stagedUploadID string,
	fromKey string,
	toKey string,
	checksum string,
	status string,
	attempts int,
	nextAttemptAt time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) (*ReconciliationEvent, error) {
	event, err := NewReconciliationEvent(
		id, attachmentID, stagedUploadID, fromKey, toKey, checksum, createdAt,
	)
	if err != nil {
		return nil, err
	}
	eventStatus := ReconciliationStatus(status)
	if !eventStatus.IsValid() {
		return nil, ErrInvalidReconciliationState
	}
	if attempts < 0 || nextAttemptAt.IsZero() || updatedAt.IsZero() {
		return nil, ErrInvalidReconciliationEvent
	}
	event.status = eventStatus
	event.attempts = attempts
	event.nextAttemptAt = nextAttemptAt
	event.updatedAt = updatedAt
	return event, nil
}

func isValidReconciliationID(id string) bool {
	return len(id) > len(ReconciliationEventIDPrefix)+1 &&
		id[:len(ReconciliationEventIDPrefix)+1] == ReconciliationEventIDPrefix+"-"
}

// FailureSchedule decides what a failed reconciliation attempt writes back.
// Retryable failures keep the attachment PROCESSING (never downloadable) and
// move the event to FAILED with exponential backoff. Once maxAttempts is
// reached, both the event and the attachment are terminal.
func (e *ReconciliationEvent) FailureSchedule(
	maxAttempts int,
	backoff time.Duration,
	now time.Time,
) (ReconciliationStatus, int, time.Time) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	if backoff <= 0 {
		backoff = time.Second
	}

	attempts := e.attempts + 1
	if attempts >= maxAttempts {
		return ReconciliationStatusDead, attempts, now.UTC()
	}

	// Cap the exponent so a long-running deployment cannot overflow the
	// duration even if maxAttempts is configured very high.
	exponent := attempts - 1
	if exponent > 10 {
		exponent = 10
	}
	delay := backoff * time.Duration(1<<uint(exponent))
	if jitterMax := delay / 4; jitterMax > 0 {
		delay += time.Duration(rand.Int64N(int64(jitterMax)))
	}
	next := now.UTC().Add(delay)
	return ReconciliationStatusFailed, attempts, next
}

func (e *ReconciliationEvent) ID() string                   { return e.id }
func (e *ReconciliationEvent) AttachmentID() string         { return e.attachmentID }
func (e *ReconciliationEvent) StagedUploadID() string       { return e.stagedUploadID }
func (e *ReconciliationEvent) FromKey() string              { return e.fromKey }
func (e *ReconciliationEvent) ToKey() string                { return e.toKey }
func (e *ReconciliationEvent) Checksum() string             { return e.checksum }
func (e *ReconciliationEvent) Status() ReconciliationStatus { return e.status }
func (e *ReconciliationEvent) Attempts() int                { return e.attempts }
func (e *ReconciliationEvent) NextAttemptAt() time.Time     { return e.nextAttemptAt }
func (e *ReconciliationEvent) CreatedAt() time.Time         { return e.createdAt }
func (e *ReconciliationEvent) UpdatedAt() time.Time         { return e.updatedAt }

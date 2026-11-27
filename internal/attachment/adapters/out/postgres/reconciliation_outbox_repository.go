package attachmentpostgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/postgres/models"
	attachmentdomain "github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/lib/pq"
)

var _ attachmentout.ReconciliationOutboxRepository = (*ReconciliationOutboxRepository)(nil)

type ReconciliationOutboxRepository struct {
	repository db.Repository
}

func NewReconciliationOutboxRepository(repository db.Repository) *ReconciliationOutboxRepository {
	return &ReconciliationOutboxRepository{repository: repository}
}

var saveReconciliationEventQuery = `
	INSERT INTO attachment_reconciliation_outbox (
		id, attachment_id, staged_upload_id, from_key, to_key, checksum,
		status, attempts, next_attempt_at, created_at, updated_at
	)
	VALUES (
		:id, :attachment_id, :staged_upload_id, :from_key, :to_key, :checksum,
		:status, :attempts, :next_attempt_at, :created_at, :updated_at
	)
`

func (r *ReconciliationOutboxRepository) Save(
	ctx context.Context,
	event *attachmentdomain.ReconciliationEvent,
) error {
	_, err := r.repository.Executor(ctx).NamedExecContext(
		ctx, saveReconciliationEventQuery, models.ReconciliationOutboxFromDomain(event),
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && string(pqErr.Code) == pqUniqueViolationCode {
			return attachmentout.ErrReconciliationEventExists
		}
		return err
	}
	return nil
}

var getReconciliationEventColumns = `
	id, attachment_id, staged_upload_id, from_key, to_key,
	TRIM(checksum) AS checksum, status, attempts,
	next_attempt_at, created_at, updated_at
`

var getReconciliationEventByIDQuery = `
	SELECT ` + getReconciliationEventColumns + `
	FROM attachment_reconciliation_outbox
	WHERE id = $1
`

// GetByID is part of the outbox port; all SQL for the outbox table stays in
// this adapter.
func (r *ReconciliationOutboxRepository) GetByID(
	ctx context.Context,
	id string,
) (*attachmentdomain.ReconciliationEvent, error) {
	var model models.ReconciliationOutboxModel
	if err := r.repository.Executor(ctx).GetContext(ctx, &model, getReconciliationEventByIDQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, attachmentout.ErrReconciliationEventNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}

var getReconciliationEventByAttachmentIDQuery = `
	SELECT ` + getReconciliationEventColumns + `
	FROM attachment_reconciliation_outbox
	WHERE attachment_id = $1
`

func (r *ReconciliationOutboxRepository) GetByAttachmentID(
	ctx context.Context,
	attachmentID string,
) (*attachmentdomain.ReconciliationEvent, error) {
	var model models.ReconciliationOutboxModel
	if err := r.repository.Executor(ctx).GetContext(
		ctx, &model, getReconciliationEventByAttachmentIDQuery, attachmentID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, attachmentout.ErrReconciliationEventNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}

var claimDueReconciliationEventQuery = `
	WITH due AS (
		SELECT id
		FROM attachment_reconciliation_outbox
		WHERE status IN ('pending', 'failed') AND next_attempt_at <= $1
		ORDER BY next_attempt_at, id
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	)
	UPDATE attachment_reconciliation_outbox AS e
	SET status = 'processing', updated_at = $1
	FROM due
	WHERE e.id = due.id
	RETURNING ` + getReconciliationEventColumnsWithAlias("e")

func getReconciliationEventColumnsWithAlias(alias string) string {
	return `
		` + alias + `.id, ` + alias + `.attachment_id, ` + alias + `.staged_upload_id,
		` + alias + `.from_key, ` + alias + `.to_key,
		TRIM(` + alias + `.checksum) AS checksum, ` + alias + `.status,
		` + alias + `.attempts, ` + alias + `.next_attempt_at,
		` + alias + `.created_at, ` + alias + `.updated_at
	`
}

// ClaimDue takes one row without waiting for another worker. It is intended to
// run inside the same transaction as the storage move so a crash returns the
// event to its previous state rather than leaving it processing forever.
func (r *ReconciliationOutboxRepository) ClaimDue(
	ctx context.Context,
	now time.Time,
) (*attachmentdomain.ReconciliationEvent, error) {
	var model models.ReconciliationOutboxModel
	if err := r.repository.Executor(ctx).GetContext(
		ctx, &model, claimDueReconciliationEventQuery, now.UTC(),
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, attachmentout.ErrNoDueEvents
		}
		return nil, err
	}
	return model.ToDomain()
}

var markReconciliationEventDoneQuery = `
	UPDATE attachment_reconciliation_outbox
	SET status = 'done', updated_at = $2
	WHERE id = $1 AND status = 'processing'
`

func (r *ReconciliationOutboxRepository) MarkDone(
	ctx context.Context,
	id string,
	now time.Time,
) error {
	return r.updateOwnedEvent(ctx, markReconciliationEventDoneQuery, id, now)
}

var recordReconciliationEventFailureQuery = `
	UPDATE attachment_reconciliation_outbox
	SET status = $2, attempts = $3, next_attempt_at = $4, updated_at = $5
	WHERE id = $1 AND status = 'processing'
`

func (r *ReconciliationOutboxRepository) RecordFailure(
	ctx context.Context,
	id string,
	status attachmentdomain.ReconciliationStatus,
	attempts int,
	nextAttemptAt time.Time,
	now time.Time,
) error {
	result, err := r.repository.Executor(ctx).ExecContext(
		ctx, recordReconciliationEventFailureQuery,
		id, status.String(), attempts, nextAttemptAt.UTC(), now.UTC(),
	)
	if err != nil {
		return err
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if updated == 0 {
		return attachmentout.ErrReconciliationEventConflict
	}
	return nil
}

func (r *ReconciliationOutboxRepository) updateOwnedEvent(
	ctx context.Context,
	query string,
	id string,
	now time.Time,
) error {
	result, err := r.repository.Executor(ctx).ExecContext(ctx, query, id, now.UTC())
	if err != nil {
		return err
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if updated == 0 {
		return attachmentout.ErrReconciliationEventConflict
	}
	return nil
}

var requeueOutboxByAttachmentIDsQuery = `
	UPDATE attachment_reconciliation_outbox
	SET status = 'pending', next_attempt_at = $2, updated_at = $2
	WHERE attachment_id = ANY($1::varchar[]) AND status IN ('processing', 'failed')
`

func (r *ReconciliationOutboxRepository) RequeueByAttachmentIDs(
	ctx context.Context,
	attachmentIDs []string,
	now time.Time,
) (int64, error) {
	if len(attachmentIDs) == 0 {
		return 0, nil
	}
	res, err := r.repository.Executor(ctx).ExecContext(
		ctx, requeueOutboxByAttachmentIDsQuery, pq.Array(attachmentIDs), now.UTC(),
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

var countOutboxDepthQuery = `
	SELECT
		COUNT(*) FILTER (WHERE status IN ('pending', 'failed', 'processing')) AS active,
		COUNT(*) FILTER (WHERE status = 'dead') AS dead
	FROM attachment_reconciliation_outbox
`

func (r *ReconciliationOutboxRepository) CountDepth(ctx context.Context) (int64, int64, error) {
	var row struct {
		Active int64 `db:"active"`
		Dead   int64 `db:"dead"`
	}
	if err := r.repository.Executor(ctx).GetContext(ctx, &row, countOutboxDepthQuery); err != nil {
		return 0, 0, err
	}
	return row.Active, row.Dead, nil
}

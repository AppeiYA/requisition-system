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
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/lib/pq"
)

var (
	ErrAttachmentNotFoundRepo = customerrors.NewNotFoundError("attachment not found")
	ErrAttachmentExistsRepo   = attachmentout.ErrAttachmentExists
)

const pqUniqueViolationCode = "23505"

type AttachmentRepository struct {
	repository db.Repository
}

func NewAttachmentRepository(repository db.Repository) *AttachmentRepository {
	return &AttachmentRepository{repository: repository}
}

var _ attachmentout.AttachmentRepository = (*AttachmentRepository)(nil)

var saveAttachmentQuery = `
	INSERT INTO attachments (
		id, ref_name, ref_id, storage_key, file_name, content_type,
		file_size, checksum, uploaded_by, created_at, status, staged_upload_id
	)
	VALUES (
		:id, :ref_name, :ref_id, :storage_key, :file_name, :content_type,
		:file_size, :checksum, :uploaded_by, :created_at, :status, :staged_upload_id
	)
`

func (r *AttachmentRepository) Save(ctx context.Context, attachment *attachmentdomain.Attachment) error {
	_, err := r.repository.Executor(ctx).NamedExecContext(ctx, saveAttachmentQuery, models.AttachmentFromDomain(attachment))
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && string(pqErr.Code) == pqUniqueViolationCode {
			return attachmentout.ErrAttachmentExists
		}
		return err
	}
	return nil
}

var updateAttachmentStatusQuery = `
	UPDATE attachments SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
`

// UpdateStatus persists only the lifecycle column. The use case calls the
// aggregate's transition methods first, so this adapter never chooses a
// status; it only refuses to write one for a row that has disappeared.
func (r *AttachmentRepository) UpdateStatus(
	ctx context.Context,
	id string,
	status attachmentdomain.AttachmentStatus,
) error {
	result, err := r.repository.Executor(ctx).ExecContext(
		ctx, updateAttachmentStatusQuery, status.String(), id,
	)
	if err != nil {
		return err
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if updated == 0 {
		return ErrAttachmentNotFoundRepo
	}
	return nil
}

var getAttachmentByIDQuery = `
	SELECT id, ref_name, ref_id, storage_key, file_name, content_type,
	       file_size, TRIM(checksum) AS checksum, uploaded_by, created_at,
	       status, staged_upload_id
	FROM attachments
	WHERE id = $1
`

func (r *AttachmentRepository) GetByID(ctx context.Context, id string) (*attachmentdomain.Attachment, error) {
	var model models.AttachmentModel
	if err := r.repository.Executor(ctx).GetContext(ctx, &model, getAttachmentByIDQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAttachmentNotFoundRepo
		}
		return nil, err
	}
	return model.ToDomain()
}

var listAttachmentsByRefQuery = `
	SELECT id, ref_name, ref_id, storage_key, file_name, content_type,
	       file_size, TRIM(checksum) AS checksum, uploaded_by, created_at,
	       status, staged_upload_id
	FROM attachments
	WHERE ref_name = $1 AND ref_id = $2
	ORDER BY created_at DESC, id DESC
	LIMIT $3 OFFSET $4
`

var countAttachmentsByRefQuery = `
	SELECT COUNT(*) FROM attachments
	WHERE ref_name = $1 AND ref_id = $2
`

func (r *AttachmentRepository) ListByRef(
	ctx context.Context,
	refName, refID string,
	limit, offset int,
) ([]*attachmentdomain.Attachment, int, error) {
	executor := r.repository.Executor(ctx)

	var modelsList []models.AttachmentModel
	if err := executor.SelectContext(ctx, &modelsList, listAttachmentsByRefQuery, refName, refID, limit, offset); err != nil {
		return nil, 0, err
	}

	var total int
	if err := executor.GetContext(ctx, &total, countAttachmentsByRefQuery, refName, refID); err != nil {
		return nil, 0, err
	}

	attachments := make([]*attachmentdomain.Attachment, 0, len(modelsList))
	for _, model := range modelsList {
		attachment, err := model.ToDomain()
		if err != nil {
			return nil, 0, err
		}
		attachments = append(attachments, attachment)
	}
	return attachments, total, nil
}

func (r *AttachmentRepository) CountByRef(ctx context.Context, refName, refID string) (int, error) {
	var total int
	if err := r.repository.Executor(ctx).GetContext(ctx, &total, countAttachmentsByRefQuery, refName, refID); err != nil {
		return 0, err
	}
	return total, nil
}

var deleteAttachmentQuery = `
	DELETE FROM attachments WHERE id = $1
`

func (r *AttachmentRepository) Delete(ctx context.Context, id string) error {
	result, err := r.repository.Executor(ctx).ExecContext(ctx, deleteAttachmentQuery, id)
	if err != nil {
		return err
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if deleted == 0 {
		return ErrAttachmentNotFoundRepo
	}
	return nil
}

var markProcessingAttachmentsDeletingByRefQuery = `
	UPDATE attachments
	SET status = 'deleting', updated_at = CURRENT_TIMESTAMP
	WHERE ref_name = $1 AND ref_id = $2 AND status = 'processing'
`

var deleteNonDeletingAttachmentsByRefQuery = `
	DELETE FROM attachments
	WHERE ref_name = $1 AND ref_id = $2 AND status <> 'deleting'
	RETURNING storage_key
`

// DeleteByRef marks in-flight processing rows as deleting (so the worker
// cleans up any moved object) and deletes all other rows, returning their
// storage keys so the caller can delete the stored objects after commit.
func (r *AttachmentRepository) DeleteByRef(ctx context.Context, refName, refID string) ([]string, error) {
	executor := r.repository.Executor(ctx)

	if _, err := executor.ExecContext(ctx, markProcessingAttachmentsDeletingByRefQuery, refName, refID); err != nil {
		return nil, err
	}

	var storageKeys []string
	if err := executor.SelectContext(ctx, &storageKeys, deleteNonDeletingAttachmentsByRefQuery, refName, refID); err != nil {
		return nil, err
	}
	if storageKeys == nil {
		return []string{}, nil
	}
	return storageKeys, nil
}

var resetStuckProcessingAttachmentsQuery = `
	WITH stuck AS (
		SELECT id
		FROM attachments
		WHERE status = 'processing' AND updated_at <= $1
		ORDER BY updated_at, id
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	)
	UPDATE attachments AS a
	SET status = 'pending', updated_at = CURRENT_TIMESTAMP
	FROM stuck
	WHERE a.id = stuck.id AND a.status = 'processing'
	RETURNING a.id
`

func (r *AttachmentRepository) ResetStuckProcessing(
	ctx context.Context,
	stuckBefore time.Time,
	limit int,
) ([]string, error) {
	if limit <= 0 {
		return []string{}, nil
	}
	var ids []string
	if err := r.repository.Executor(ctx).SelectContext(
		ctx, &ids, resetStuckProcessingAttachmentsQuery, stuckBefore.UTC(), limit,
	); err != nil {
		return nil, err
	}
	if ids == nil {
		return []string{}, nil
	}
	return ids, nil
}

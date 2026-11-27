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
	ErrStagedUploadNotFoundRepo = customerrors.NewNotFoundError("staged upload not found")
	ErrStagedUploadExistsRepo   = customerrors.NewConflictError("staged upload already exists")
)

const stagedUploadColumns = `
	id, uploaded_by, storage_path, original_filename, content_type,
	file_size, TRIM(checksum) AS checksum, status, created_at, expires_at, claimed_at
`

var saveStagedUploadQuery = `
	INSERT INTO staged_uploads (
		id, uploaded_by, storage_path, original_filename, content_type,
		file_size, checksum, status, created_at, expires_at
	)
	VALUES (
		:id, :uploaded_by, :storage_path, :original_filename, :content_type,
		:file_size, :checksum, :status, :created_at, :expires_at
	)
`

type StagedUploadRepository struct {
	repository db.Repository
}

func NewStagedUploadRepository(repository db.Repository) *StagedUploadRepository {
	return &StagedUploadRepository{repository: repository}
}

var _ attachmentout.StagedUploadRepository = (*StagedUploadRepository)(nil)

func (r *StagedUploadRepository) Save(ctx context.Context, upload *attachmentdomain.StagedUpload) error {
	_, err := r.repository.Executor(ctx).NamedExecContext(
		ctx, saveStagedUploadQuery, models.StagedUploadFromDomain(upload),
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && string(pqErr.Code) == pqUniqueViolationCode {
			return ErrStagedUploadExistsRepo
		}
		return err
	}
	return nil
}

var getStagedUploadByIDQuery = `
	SELECT ` + stagedUploadColumns + `
	FROM staged_uploads
	WHERE id = $1
`

func (r *StagedUploadRepository) GetByID(ctx context.Context, id string) (*attachmentdomain.StagedUpload, error) {
	var model models.StagedUploadModel
	if err := r.repository.Executor(ctx).GetContext(ctx, &model, getStagedUploadByIDQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStagedUploadNotFoundRepo
		}
		return nil, err
	}
	return model.ToDomain()
}

var claimStagedUploadsQuery = `
	UPDATE staged_uploads
	SET status = 'claimed', claimed_at = $3
	WHERE id = ANY($1::varchar[])
	  AND status = 'pending'
	  AND uploaded_by = $2
	  AND expires_at > $3
	RETURNING ` + stagedUploadColumns + `
`

// ClaimByIDs atomically claims all matching pending, owned, unexpired staged
// uploads in one statement. The caller verifies len(claimed) == len(ids) and
// rolls back the transaction on any mismatch.
func (r *StagedUploadRepository) ClaimByIDs(
	ctx context.Context,
	ids []string,
	actorID string,
	now time.Time,
) ([]*attachmentdomain.StagedUpload, error) {
	if len(ids) == 0 {
		return []*attachmentdomain.StagedUpload{}, nil
	}

	var rows []models.StagedUploadModel
	if err := r.repository.Executor(ctx).SelectContext(
		ctx, &rows, claimStagedUploadsQuery, pq.Array(ids), actorID, now.UTC(),
	); err != nil {
		return nil, err
	}

	uploads := make([]*attachmentdomain.StagedUpload, 0, len(rows))
	for _, row := range rows {
		upload, err := row.ToDomain()
		if err != nil {
			return nil, err
		}
		uploads = append(uploads, upload)
	}
	return uploads, nil
}

var countPendingStagedByUserQuery = `
	SELECT COUNT(*), COALESCE(SUM(file_size), 0)
	FROM staged_uploads
	WHERE uploaded_by = $1 AND status = 'pending' AND expires_at > $2
`

func (r *StagedUploadRepository) CountPendingByUser(
	ctx context.Context,
	actorID string,
	now time.Time,
) (int, int64, error) {
	var row struct {
		Count int   `db:"count"`
		Bytes int64 `db:"coalesce"`
	}
	if err := r.repository.Executor(ctx).GetContext(
		ctx, &row, countPendingStagedByUserQuery, actorID, now.UTC(),
	); err != nil {
		return 0, 0, err
	}
	return row.Count, row.Bytes, nil
}

var expirePendingStagedUploadsQuery = `
	WITH expired AS (
		SELECT id
		FROM staged_uploads
		WHERE status = 'pending' AND expires_at <= $1
		ORDER BY expires_at, id
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	)
	UPDATE staged_uploads AS s
	SET status = 'expired'
	FROM expired
	WHERE s.id = expired.id AND s.status = 'pending'
	RETURNING
		s.id, s.uploaded_by, s.storage_path, s.original_filename,
		s.content_type, s.file_size, TRIM(s.checksum) AS checksum,
		s.status, s.created_at, s.expires_at, s.claimed_at
`

// ExpirePending transitions only pending rows past expires_at to expired.
// Claimed rows are never touched.
func (r *StagedUploadRepository) ExpirePending(
	ctx context.Context,
	now time.Time,
	limit int,
) ([]*attachmentdomain.StagedUpload, error) {
	if limit <= 0 {
		return []*attachmentdomain.StagedUpload{}, nil
	}
	var rows []models.StagedUploadModel
	if err := r.repository.Executor(ctx).SelectContext(
		ctx, &rows, expirePendingStagedUploadsQuery, now.UTC(), limit,
	); err != nil {
		return nil, err
	}
	uploads := make([]*attachmentdomain.StagedUpload, 0, len(rows))
	for _, row := range rows {
		upload, err := row.ToDomain()
		if err != nil {
			return nil, err
		}
		uploads = append(uploads, upload)
	}
	return uploads, nil
}

var countAllPendingStagedQuery = `
	SELECT COUNT(*) FROM staged_uploads
	WHERE status = 'pending' AND expires_at > $1
`

func (r *StagedUploadRepository) CountPendingAll(ctx context.Context, now time.Time) (int64, error) {
	var count int64
	if err := r.repository.Executor(ctx).GetContext(ctx, &count, countAllPendingStagedQuery, now.UTC()); err != nil {
		return 0, err
	}
	return count, nil
}

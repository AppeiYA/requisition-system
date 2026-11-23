package invitationpostgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/AppeiYA/requisition-system/internal/invitation/adapters/out/postgres/models"
	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/lib/pq"
)

// ErrInvitationRotationConflict is returned when the row changed between the
// state check and the token rotation. The HTTP layer maps it to 409.
var ErrInvitationRotationConflict = customerrors.NewConflictError("invitation can no longer be resent")

const invitationSummaryColumns = `
	i.id, i.first_name, i.last_name, i.email, i.department_id,
	COALESCE(d.name, '') AS department_name,
	i.job_title, i.status, i.invited_by,
	TRIM(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')) AS inviter_name,
	i.expires_at, i.accepted_at, i.revoked_at,
	i.delivered_at, i.delivery_failed_at, i.created_at
`

const invitationSummaryJoins = `
	FROM invitations i
	LEFT JOIN departments d ON d.id = i.department_id
	LEFT JOIN users u ON u.id = i.invited_by
`

const invitationListWhere = `
	WHERE ($1 = '' OR LOWER(i.email) LIKE '%' || LOWER($1) || '%'
		OR LOWER(i.first_name) LIKE '%' || LOWER($1) || '%'
		OR LOWER(i.last_name) LIKE '%' || LOWER($1) || '%')
	  AND ($2 = '' OR i.status = $2)
	  AND ($3 = '' OR i.department_id = $3)
	  AND ($4 = '' OR i.invited_by = $4)
	  AND ($5::TIMESTAMPTZ IS NULL OR i.created_at >= $5)
	  AND ($6::TIMESTAMPTZ IS NULL OR i.created_at < $6)
`

func invitationFilterArgs(filter invitationdomain.InvitationListFilter) []any {
	return []any{
		filter.Search,
		filter.Status,
		filter.DepartmentID,
		filter.InvitedBy,
		filter.From,
		filter.To,
	}
}

func (r *InvitationRepository) ListInvitations(ctx context.Context, filter invitationdomain.InvitationListFilter) ([]invitationdomain.InvitationSummary, error) {
	query := `SELECT` + invitationSummaryColumns + invitationSummaryJoins + invitationListWhere + `
	ORDER BY i.created_at DESC, i.id DESC
	LIMIT $7 OFFSET $8`

	args := append(invitationFilterArgs(filter), filter.Limit, filter.Offset)
	var rows []models.InvitationSummaryModel
	if err := r.repository.Executor(ctx).SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	summaries := make([]invitationdomain.InvitationSummary, len(rows))
	for i, row := range rows {
		summaries[i] = row.ToDomain()
	}
	return summaries, nil
}

func (r *InvitationRepository) CountInvitations(ctx context.Context, filter invitationdomain.InvitationListFilter) (int, error) {
	query := `SELECT COUNT(*)` + invitationSummaryJoins + invitationListWhere
	var total int
	err := r.repository.Executor(ctx).GetContext(ctx, &total, query, invitationFilterArgs(filter)...)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *InvitationRepository) GetInvitationSummary(ctx context.Context, invitationID string) (*invitationdomain.InvitationSummary, error) {
	query := `SELECT` + invitationSummaryColumns + invitationSummaryJoins + ` WHERE i.id = $1`
	var model models.InvitationSummaryModel
	if err := r.repository.Executor(ctx).GetContext(ctx, &model, query, invitationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvitationNotFoundRepo
		}
		return nil, err
	}
	summary := model.ToDomain()
	return &summary, nil
}

var lockInvitationByIDQuery = `
	SELECT id, first_name, last_name, phone_number, staff_id, department_id,
	       job_title, email, token_hash, invited_by, status,
	       expires_at, accepted_at, revoked_at, created_at
	FROM invitations
	WHERE id = $1
	FOR UPDATE
`

func (r *InvitationRepository) LockInvitationByID(ctx context.Context, invitationID string) (*invitationdomain.Invitation, error) {
	var model models.InvitationModel
	err := r.repository.Executor(ctx).GetContext(ctx, &model, lockInvitationByIDQuery, invitationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvitationNotFoundRepo
		}
		return nil, err
	}
	return model.ToDomain()
}

var rotateInvitationTokenQuery = `
	UPDATE invitations
	SET token_hash = $2,
	    expires_at = $3,
	    status = 'pending',
	    delivered_at = NULL,
	    delivery_failed_at = NULL
	WHERE id = $1
	  AND status IN ('pending', 'delivered', 'failed', 'expired')
	  AND accepted_at IS NULL
	  AND revoked_at IS NULL
`

func (r *InvitationRepository) RotateInvitationToken(ctx context.Context, invitationID, tokenHash string, expiresAt time.Time) error {
	res, err := r.repository.Executor(ctx).ExecContext(ctx, rotateInvitationTokenQuery, invitationID, tokenHash, expiresAt)
	if err != nil {
		// The partial unique index on pending/delivered emails protects against
		// two active invitations for the same address.
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && string(pqErr.Code) == "23505" {
			return ErrInvitationRotationConflict
		}
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrInvitationRotationConflict
	}
	return nil
}

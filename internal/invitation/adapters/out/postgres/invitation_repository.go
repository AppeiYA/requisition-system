package invitationpostgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AppeiYA/requisition-system/internal/invitation/adapters/out/postgres/models"
	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrInvitationNotFoundRepo = customerrors.NewNotFoundError("invitation not found")
)

type InvitationRepository struct {
	repository db.Repository
}

func NewInvitationRepository(repository db.Repository) *InvitationRepository {
	return &InvitationRepository{repository: repository}
}

var saveInvitationQuery = `
	INSERT INTO invitations (id, email, token_hash, invited_by, expires_at, accepted_at, revoked_at, created_at)
	VALUES (:id, :email, :token_hash, :invited_by, :expires_at, :accepted_at, :revoked_at, :created_at)
`

func (r *InvitationRepository) SaveInvitation(ctx context.Context, invitation *invitationdomain.Invitation) error {
	_, err := r.repository.Executor(ctx).NamedExecContext(ctx, saveInvitationQuery, models.FromDomain(invitation))
	if err != nil {
		return err
	}
	return nil
}

var getByTokenHashQuery = `
	SELECT id, email, token_hash, invited_by, expires_at, accepted_at, revoked_at, created_at
	FROM invitations
	WHERE token_hash = $1
`

func (r *InvitationRepository) GetInvitationByTokenHash(ctx context.Context, tokenHash string) (*invitationdomain.Invitation, error) {
	var model models.InvitationModel
	err := r.repository.Executor(ctx).GetContext(ctx, &model, getByTokenHashQuery, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvitationNotFoundRepo
		}
		return nil, err
	}
	return model.ToDomain()
}

var getActiveByEmailQuery = `
	SELECT id, email, token_hash, invited_by, expires_at, accepted_at, revoked_at, created_at
	FROM invitations
	WHERE LOWER(email) = LOWER($1)
	  AND accepted_at IS NULL
	  AND revoked_at IS NULL
`

func (r *InvitationRepository) GetActiveInvitationByEmail(ctx context.Context, email string) (*invitationdomain.Invitation, error) {
	var model models.InvitationModel
	err := r.repository.Executor(ctx).GetContext(ctx, &model, getActiveByEmailQuery, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvitationNotFoundRepo
		}
		return nil, err
	}
	return model.ToDomain()
}

var markAcceptedQuery = `
	UPDATE invitations
	SET accepted_at = NOW()
	WHERE id = $1 AND accepted_at IS NULL AND revoked_at IS NULL
`

func (r *InvitationRepository) MarkInvitationAccepted(ctx context.Context, invitationID string) error {
	res, err := r.repository.Executor(ctx).ExecContext(ctx, markAcceptedQuery, invitationID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrInvitationNotFoundRepo
	}
	return nil
}

var revokeQuery = `
	UPDATE invitations
	SET revoked_at = NOW()
	WHERE id = $1
`

func (r *InvitationRepository) RevokeInvitation(ctx context.Context, invitationID string) error {
	_, err := r.repository.Executor(ctx).ExecContext(ctx, revokeQuery, invitationID)
	if err != nil {
		return err
	}
	return nil
}

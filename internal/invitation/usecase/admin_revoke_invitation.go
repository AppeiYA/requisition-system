package usecase

import (
	"context"
	"strings"

	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/audit"
)

type AdminRevokeInvitationUsecase struct {
	repo      invitationout.InvitationRepository
	txManager invitationout.TransactionManager
	audit     invitationout.AuditRecorder
}

func NewAdminRevokeInvitationUsecase(
	repo invitationout.InvitationRepository,
	txManager invitationout.TransactionManager,
	auditRecorder invitationout.AuditRecorder,
) *AdminRevokeInvitationUsecase {
	return &AdminRevokeInvitationUsecase{
		repo:      repo,
		txManager: txManager,
		audit:     auditRecorder,
	}
}

func (uc *AdminRevokeInvitationUsecase) Execute(ctx context.Context, cmd invitationin.RevokeInvitationCommand) error {
	if cmd.InvitationID == "" {
		return ErrInvitationIDRequired
	}
	if cmd.ActorID == "" {
		return ErrActorIDRequired
	}
	reason := strings.TrimSpace(cmd.Reason)
	if len([]rune(reason)) > 500 {
		return ErrReasonTooLong
	}

	return uc.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		invitation, err := uc.repo.LockInvitationByID(txCtx, cmd.InvitationID)
		if err != nil {
			return err
		}
		if invitation == nil {
			return invitationdomain.ErrInvitationNotFound
		}

		switch invitation.Status() {
		case invitationdomain.InvitationStatusAccepted:
			return ErrInvitationCannotRevoke
		case invitationdomain.InvitationStatusRevoked:
			// Revocation is a terminal state; replaying the request is a
			// successful no-op rather than an error.
			return nil
		case invitationdomain.InvitationStatusPending, invitationdomain.InvitationStatusDelivered:
			// These are the only states the repository can revoke.
		default:
			return ErrInvitationNotRevocable
		}

		beforeStatus := invitation.Status().String()
		if err := invitation.MarkAsRevoked(); err != nil {
			return err
		}
		if err := uc.repo.RevokeInvitation(txCtx, invitation.ID()); err != nil {
			return err
		}
		after, err := uc.repo.GetInvitationSummary(txCtx, invitation.ID())
		if err != nil {
			return err
		}

		return uc.audit.Record(txCtx, audit.Entry{
			ActorID:    cmd.ActorID,
			Action:     "invitation.revoke",
			TargetType: "invitation",
			TargetID:   invitation.ID(),
			Before: map[string]any{
				"status":     beforeStatus,
				"revoked_at": nil,
			},
			After: map[string]any{
				"status":     after.Status,
				"revoked_at": after.RevokedAt,
			},
			Reason: reason,
		})
	})
}

package usecase

import (
	"context"
	"errors"
	"net/http"
	"time"

	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/audit"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

type AdminResendInvitationUsecase struct {
	repo      invitationout.InvitationRepository
	txManager invitationout.TransactionManager
	audit     invitationout.AuditRecorder
}

func NewAdminResendInvitationUsecase(
	repo invitationout.InvitationRepository,
	txManager invitationout.TransactionManager,
	auditRecorder invitationout.AuditRecorder,
) *AdminResendInvitationUsecase {
	return &AdminResendInvitationUsecase{
		repo:      repo,
		txManager: txManager,
		audit:     auditRecorder,
	}
}

func (uc *AdminResendInvitationUsecase) Execute(ctx context.Context, cmd invitationin.ResendInvitationCommand) (*invitationin.ResendInvitationResult, error) {
	if cmd.InvitationID == "" {
		return nil, ErrInvitationIDRequired
	}
	if cmd.ActorID == "" {
		return nil, ErrActorIDRequired
	}

	var result *invitationin.ResendInvitationResult
	if err := uc.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		invitation, err := uc.repo.LockInvitationByID(txCtx, cmd.InvitationID)
		if err != nil {
			return err
		}
		if invitation == nil {
			return invitationdomain.ErrInvitationNotFound
		}
		if invitation.Status() == invitationdomain.InvitationStatusAccepted ||
			invitation.Status() == invitationdomain.InvitationStatusRevoked {
			return ErrInvitationCannotResend
		}

		// The partial unique index forbids a second pending/delivered row for
		// the same email, so rotating a failed or expired invitation back to
		// pending must first rule out another active invitation.
		active, err := uc.repo.GetActiveInvitationByEmail(txCtx, invitation.Email().Value())
		if err != nil {
			var customErr *customerrors.CustomError
			if !errors.As(err, &customErr) || customErr.Code != http.StatusNotFound {
				return err
			}
			active = nil
		}
		if active != nil && active.ID() != invitation.ID() {
			return ErrInvitationHasActive
		}

		before, err := uc.repo.GetInvitationSummary(txCtx, invitation.ID())
		if err != nil {
			return err
		}

		token, err := generateInvitationToken()
		if err != nil {
			return err
		}
		expiresAt := time.Now().Add(invitationdomain.InvitationExpirationDuration)
		if err := uc.repo.RotateInvitationToken(
			txCtx,
			invitation.ID(),
			invitationdomain.HashInvitationToken(token),
			expiresAt,
		); err != nil {
			return err
		}

		result = &invitationin.ResendInvitationResult{
			ID:           invitation.ID(),
			Token:        token,
			Email:        invitation.Email().Value(),
			FirstName:    invitation.FirstName().Value(),
			LastName:     invitation.LastName().Value(),
			InvitedBy:    invitation.InvitedBy(),
			DepartmentID: invitation.DepartmentID(),
			JobTitle:     invitation.JobTitle(),
			ExpiresAt:    expiresAt,
		}

		return uc.audit.Record(txCtx, audit.Entry{
			ActorID:    cmd.ActorID,
			Action:     "invitation.resend",
			TargetType: "invitation",
			TargetID:   invitation.ID(),
			Before: map[string]any{
				"status":             before.Status,
				"expires_at":         before.ExpiresAt,
				"delivered_at":       before.DeliveredAt,
				"delivery_failed_at": before.DeliveryFailedAt,
			},
			After: map[string]any{
				"status":             invitationdomain.InvitationStatusPending.String(),
				"expires_at":         expiresAt,
				"delivered_at":       nil,
				"delivery_failed_at": nil,
			},
		})
	}); err != nil {
		return nil, err
	}
	return result, nil
}

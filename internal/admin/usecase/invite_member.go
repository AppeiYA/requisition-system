package usecase

import (
	"context"
	"time"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrUserAlreadyExists = customerrors.NewConflictError("user with this email already exists")
	ErrPendingInvitation = customerrors.NewConflictError("pending invitation already exists for this email")
)

type InviteMemberUsecase struct {
	emailSender             adminout.EmailSender
	adminIdentityPort       adminout.AdminIdentityPort
	adminInvitationPort     adminout.AdminInvitationPort
	adminTransactionManager adminout.TransactionManager
}

func NewInviteMemberUsecase(
	emailSender adminout.EmailSender,
	adminIdentityPort adminout.AdminIdentityPort,
	adminInvitationPort adminout.AdminInvitationPort,
	adminTransactionManager adminout.TransactionManager,
) *InviteMemberUsecase {
	return &InviteMemberUsecase{
		emailSender:             emailSender,
		adminIdentityPort:       adminIdentityPort,
		adminInvitationPort:     adminInvitationPort,
		adminTransactionManager: adminTransactionManager,
	}
}

func (u *InviteMemberUsecase) Execute(ctx context.Context, cmd adminin.InviteMemberCommand) (*adminin.InviteMemberResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	email, err := shareddomain.NewEmail(cmd.Email)
	if err != nil {
		return nil, err
	}

	if exists, err := u.adminIdentityPort.CheckUserExists(ctx, email); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrUserAlreadyExists
	}

	if exists, err := u.adminInvitationPort.CheckActiveInvitationExists(ctx, email); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrPendingInvitation
	}

	var saved *adminout.SavedInvitation
	err = u.adminTransactionManager.WithinTransaction(ctx, func(ctx context.Context) error {
		var saveErr error
		saved, saveErr = u.adminInvitationPort.SaveInvitation(ctx, email, cmd.InviterID)
		return saveErr
	})
	if err != nil {
		return nil, err
	}

	// Delivery happens after the invitation is durably persisted. A delivery
	// failure surfaces to the caller; the invitation remains valid and can be
	// re-sent.
	if err := u.emailSender.SendInvitation(ctx, email.Value(), cmd.InviterID, saved.Token, saved.ExpiresAt.Format(time.RFC3339)); err != nil {
		return nil, err
	}

	return &adminin.InviteMemberResult{InvitationID: saved.ID}, nil
}

package usecase

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrInvitationInvalidToken  = customerrors.NewUnauthorizedError("invalid invitation token")
	ErrInvitationExpired       = customerrors.NewForbiddenError("invitation has expired")
	ErrInvitationRevoked       = customerrors.NewForbiddenError("invitation has been revoked")
	ErrInvitationAlreadyUsed   = customerrors.NewForbiddenError("invitation has already been accepted")
	ErrInvitationNotFound      = customerrors.NewNotFoundError("invitation not found")
	ErrInvitationTokenRequired = customerrors.NewBadRequestError("invitation token is required")
)

type AcceptInvitationUsecase struct {
	invitationRepo invitationout.InvitationRepository
	identityPort   invitationout.InvitationIdentityPort
	txManager      invitationout.TransactionManager
}

func NewAcceptInvitationUsecase(invitationRepo invitationout.InvitationRepository, identityPort invitationout.InvitationIdentityPort, txManager invitationout.TransactionManager) *AcceptInvitationUsecase {
	return &AcceptInvitationUsecase{
		invitationRepo: invitationRepo,
		identityPort:   identityPort,
		txManager:      txManager,
	}
}

func (uc *AcceptInvitationUsecase) Execute(ctx context.Context, cmd invitationin.AcceptInvitationCommand) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if cmd.Token == "" {
		return ErrInvitationInvalidToken
	}

	tokenHash := domain.HashInvitationToken(cmd.Token)

	inv, err := uc.invitationRepo.GetInvitationByTokenHash(ctx, tokenHash)
	if err != nil {
		var custErr *customerrors.CustomError
		if errors.As(err, &custErr) && custErr.Code == http.StatusNotFound {
			return ErrInvitationNotFound
		}
		return err
	}

	now := time.Now()
	if inv.IsExpired(now) {
		return ErrInvitationExpired
	}
	if inv.IsRevoked() {
		return ErrInvitationRevoked
	}
	if inv.IsAccepted() {
		return ErrInvitationAlreadyUsed
	}

	err = uc.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		regCmd := &invitationout.CreateIdentityRequest{
			FirstName: cmd.FirstName,
			LastName:  cmd.LastName,
			Email:     inv.Email().Value(),
			Password:  cmd.Password,
		}

		if err := uc.identityPort.RegisterUser(ctx, regCmd); err != nil {
			return err
		}

		if err := uc.invitationRepo.MarkInvitationAccepted(ctx, inv.ID()); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

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
	departmentPort invitationout.InvitationDepartmentPort
	txManager      invitationout.TransactionManager
}

func NewAcceptInvitationUsecase(
	invitationRepo invitationout.InvitationRepository,
	identityPort invitationout.InvitationIdentityPort,
	departmentPort invitationout.InvitationDepartmentPort,
	txManager invitationout.TransactionManager,
) *AcceptInvitationUsecase {
	return &AcceptInvitationUsecase{
		invitationRepo: invitationRepo,
		identityPort:   identityPort,
		departmentPort: departmentPort,
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
		// Names are set by the admin in the invitation and cannot be changed by
		// the invitee. The invitee may override the prefilled phone number.
		phone := cmd.Phone
		if phone == "" {
			phone = inv.Phone().Value()
		}

		regCmd := &invitationout.CreateIdentityRequest{
			FirstName: inv.FirstName().Value(),
			LastName:  inv.LastName().Value(),
			Phone:     phone,
			StaffID:   inv.StaffID(),
			Email:     inv.Email().Value(),
			Password:  cmd.Password,
		}

		userID, err := uc.identityPort.RegisterUser(ctx, regCmd)
		if err != nil {
			return err
		}

		// The user and their department membership are created together: if
		// either insert fails the whole transaction rolls back.
		membershipCmd := &invitationout.CreateMembershipRequest{
			UserID:       userID,
			DepartmentID: inv.DepartmentID(),
			JobTitle:     inv.JobTitle(),
		}
		if err := uc.departmentPort.CreateMembership(ctx, membershipCmd); err != nil {
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

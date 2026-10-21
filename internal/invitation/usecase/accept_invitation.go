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
	telemetry      invitationout.Telemetry
}

func NewAcceptInvitationUsecase(
	invitationRepo invitationout.InvitationRepository,
	identityPort invitationout.InvitationIdentityPort,
	departmentPort invitationout.InvitationDepartmentPort,
	txManager invitationout.TransactionManager,
	telemetry invitationout.Telemetry,
) *AcceptInvitationUsecase {
	return &AcceptInvitationUsecase{
		invitationRepo: invitationRepo,
		identityPort:   identityPort,
		departmentPort: departmentPort,
		txManager:      txManager,
		telemetry:      telemetry,
	}
}

func (uc *AcceptInvitationUsecase) Execute(ctx context.Context, cmd invitationin.AcceptInvitationCommand) error {
	// A cancelled request never starts the operation, so it is not measured.
	if err := ctx.Err(); err != nil {
		return err
	}

	var operationErr error

	// The measurement covers the whole operation, including the transaction
	// that registers the user, creates the membership and accepts the
	// invitation. The outcome of the operation is the one reported to the
	// caller, the instrumentation can never replace nor mask it.
	_ = uc.telemetry.Measure(ctx, invitationout.OperationAcceptInvitation, func(ctx context.Context) error {
		operationErr = uc.acceptInvitation(ctx, cmd)
		return operationErr
	})

	return operationErr
}

func (uc *AcceptInvitationUsecase) acceptInvitation(ctx context.Context, cmd invitationin.AcceptInvitationCommand) error {
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
		inv, err := uc.invitationRepo.LockInvitationByTokenHash(ctx, tokenHash)
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

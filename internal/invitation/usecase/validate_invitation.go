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

type ValidateInvitationUsecase struct {
	invitationRepo invitationout.InvitationRepository
}

func NewValidateInvitationUsecase(invitationRepo invitationout.InvitationRepository) *ValidateInvitationUsecase {
	return &ValidateInvitationUsecase{invitationRepo: invitationRepo}
}

func (u *ValidateInvitationUsecase) Execute(ctx context.Context, token string) (*invitationin.ValidateInvitationResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if token == "" {
		return &invitationin.ValidateInvitationResult{Valid: false}, nil
	}

	tokenHash := domain.HashInvitationToken(token)
	inv, err := u.invitationRepo.GetInvitationByTokenHash(ctx, tokenHash)
	if err != nil {
		var custErr *customerrors.CustomError
		if errors.As(err, &custErr) && custErr.Code == http.StatusNotFound {
			return &invitationin.ValidateInvitationResult{Valid: false}, nil
		}
		return nil, err
	}

	now := time.Now()
	if inv.IsExpired(now) || inv.IsRevoked() || inv.IsAccepted() {
		return &invitationin.ValidateInvitationResult{Valid: false, Email: inv.Email().Value()}, nil
	}

	return &invitationin.ValidateInvitationResult{
		Valid:        true,
		Email:        inv.Email().Value(),
		FirstName:    inv.FirstName().Value(),
		LastName:     inv.LastName().Value(),
		Phone:        inv.Phone().Value(),
		StaffID:      inv.StaffID(),
		DepartmentID: inv.DepartmentID(),
		JobTitle:     inv.JobTitle(),
	}, nil
}

package usecase

import (
	"context"

	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

type GetInvitationUsecase struct {
	inivitationRepo invitationout.InvitationRepository
}

func NewGetInvitationUsecase(invitationRepo invitationout.InvitationRepository) *GetInvitationUsecase {
	return &GetInvitationUsecase{
		inivitationRepo: invitationRepo,
	}
}

func (uc *GetInvitationUsecase) ByEmail(ctx context.Context, email string) (*invitationin.GetInvitationOutput, error) {
	vaildEmail, err := shareddomain.NewEmail(email)
	if err != nil {
		return nil, err
	}

	inv, err := uc.inivitationRepo.GetActiveInvitationByEmail(ctx, vaildEmail.Value())
	if err != nil {
		return nil, err
	}

	return &invitationin.GetInvitationOutput{
		ID:           inv.ID(),
		FirstName:    inv.FirstName().Value(),
		LastName:     inv.LastName().Value(),
		Phone:        inv.Phone().Value(),
		StaffID:      inv.StaffID(),
		DepartmentID: inv.DepartmentID(),
		JobTitle:     inv.JobTitle(),
		Email:        inv.Email().Value(),
		InvitedBy:    inv.InvitedBy(),
		ExpiresAt:    inv.ExpiresAt(),
		AcceptedAt:   inv.AcceptedAt(),
		RevokedAt:    inv.RevokedAt(),
		CreatedAt:    inv.CreatedAt(),
	}, nil
}
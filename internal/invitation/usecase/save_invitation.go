package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

type SaveInvitationUsecase struct {
	invitationRepo invitationout.InvitationRepository
	idGenerator    invitationout.IDGenerator
}

func NewSaveInvitationUsecase(invitationRepo invitationout.InvitationRepository, idGenerator invitationout.IDGenerator) *SaveInvitationUsecase {
	return &SaveInvitationUsecase{
		invitationRepo: invitationRepo,
		idGenerator:    idGenerator,
	}
}

func generateInvitationToken() (string, error) {
	b := make([]byte, domain.InvitationTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (uc *SaveInvitationUsecase) Execute(ctx context.Context, cmd invitationin.SaveInvitationCommand) (*invitationin.SaveInvitationResponse, error) {
	validEmail, err := shareddomain.NewEmail(cmd.Email)
	if err != nil {
		return nil, err
	}

	err = uc.checkExistingInvitation(ctx, validEmail.Value())
	if err != nil {
		return nil, err
	}

	firstName, err := shareddomain.NewUsername(cmd.FirstName, shareddomain.FirstName)
	if err != nil {
		return nil, err
	}

	lastName, err := shareddomain.NewUsername(cmd.LastName, shareddomain.LastName)
	if err != nil {
		return nil, err
	}

	phone, err := shareddomain.NewOptionalPhoneNumber(cmd.Phone)
	if err != nil {
		return nil, err
	}

	token, err := generateInvitationToken()
	if err != nil {
		return nil, err
	}

	tokenHash := domain.HashInvitationToken(token)

	invitationID, err := uc.idGenerator.GenerateID(domain.InvitationIDPrefix)
	if err != nil {
		return nil, err
	}

	invitation, err := domain.NewInvitation(invitationID, firstName, lastName, phone, cmd.StaffID, cmd.DepartmentID, cmd.JobTitle, validEmail, tokenHash, cmd.InviterID)
	if err != nil {
		return nil, err
	}

	err = uc.invitationRepo.SaveInvitation(ctx, invitation)
	if err != nil {
		return nil, err
	}

	response := &invitationin.SaveInvitationResponse{
		ID:        invitationID,
		Token:     token,
		ExpiresAt: invitation.ExpiresAt(),
	}

	return response, nil
}

func (uc *SaveInvitationUsecase) checkExistingInvitation(
	ctx context.Context,
	email string,
) error {
	existingInvitation, err := uc.invitationRepo.GetInvitationByEmail(ctx, email)

	if err != nil {
		var customErr *customerrors.CustomError
		if errors.As(err, &customErr) && customErr.Code == http.StatusNotFound {
			return nil
		}

		return err
	}

	now := time.Now()

	if existingInvitation.IsExpired(now) {
		if existingInvitation.Status() == domain.InvitationStatusPending ||
			existingInvitation.Status() == domain.InvitationStatusDelivered {
			if err := uc.invitationRepo.UpdateInvitationStatus(
				ctx,
				existingInvitation.ID(),
				domain.InvitationStatusExpired,
			); err != nil {
				return err
			}
		}
		return nil
	}

	switch existingInvitation.Status() {
	case domain.InvitationStatusPending,
		domain.InvitationStatusDelivered:

		return customerrors.NewConflictError(
			"pending invitation already exists for this email",
		)

	case domain.InvitationStatusAccepted:
		return customerrors.NewConflictError(
			"invitation has already been accepted",
		)

	case domain.InvitationStatusRevoked,
		domain.InvitationStatusExpired,
		domain.InvitationStatusFailed:

		return nil

	default:
		return nil
	}
}

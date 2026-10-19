package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
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

	now := time.Now()
	expiresAt := now.Add(domain.InvitationExpirationDuration)

	invitation := domain.NewInvitation(invitationID, firstName, lastName, phone, cmd.StaffID, cmd.DepartmentID, cmd.JobTitle, validEmail, tokenHash, cmd.InviterID, expiresAt, now)

	err = uc.invitationRepo.SaveInvitation(ctx, invitation)
	if err != nil {
		return nil, err
	}

	response := &invitationin.SaveInvitationResponse{
		ID:        invitationID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	return response, nil
}

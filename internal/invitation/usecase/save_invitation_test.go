package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationmocks "github.com/AppeiYA/requisition-system/internal/invitation/mocks"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

func Test_SaveInvitationUsecase_Success(t *testing.T) {
	t.Parallel()

	var saved *domain.Invitation
	repo := invitationmocks.NewMockInvitationRepository()
	repo.SaveInvitationFunc = func(ctx context.Context, invitation *domain.Invitation) error {
		saved = invitation
		return nil
	}

	uc := NewSaveInvitationUsecase(repo, invitationmocks.NewMockIDGenerator())

	res, err := uc.Execute(context.Background(), "invitee@example.com", invitationmocks.DefaultInviterID)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if res.ID != "inv123" {
		t.Errorf("expected id %q, got %q", "inv123", res.ID)
	}
	if res.Token == "" {
		t.Fatal("expected a raw token to be returned")
	}
	if saved == nil {
		t.Fatal("expected invitation to be saved")
	}
	if saved.TokenHash() != domain.HashInvitationToken(res.Token) {
		t.Errorf("stored token hash does not match returned token")
	}
	if saved.Email().Value() != "invitee@example.com" {
		t.Errorf("expected email %q, got %q", "invitee@example.com", saved.Email().Value())
	}
	if saved.InvitedBy() != invitationmocks.DefaultInviterID {
		t.Errorf("expected inviter %q, got %q", invitationmocks.DefaultInviterID, saved.InvitedBy())
	}

	remaining := time.Until(saved.ExpiresAt())
	if remaining <= 23*time.Hour || remaining > 24*time.Hour {
		t.Errorf("expected ~24h expiry, got %v", remaining)
	}
}

func Test_SaveInvitationUsecase_InvalidEmail(t *testing.T) {
	t.Parallel()

	uc := NewSaveInvitationUsecase(invitationmocks.NewMockInvitationRepository(), invitationmocks.NewMockIDGenerator())

	_, err := uc.Execute(context.Background(), "not-an-email", invitationmocks.DefaultInviterID)
	if err != shareddomain.ErrInvalidEmailFormat {
		t.Fatalf("expected ErrInvalidEmailFormat, got: %v", err)
	}
}

func Test_SaveInvitationUsecase_IDGeneratorError(t *testing.T) {
	t.Parallel()

	idErr := errors.New("id generation failed")
	idGen := invitationmocks.NewMockIDGenerator()
	idGen.GenerateIDFunc = func(prefix string) (string, error) {
		return "", idErr
	}

	uc := NewSaveInvitationUsecase(invitationmocks.NewMockInvitationRepository(), idGen)

	_, err := uc.Execute(context.Background(), "invitee@example.com", invitationmocks.DefaultInviterID)
	if !errors.Is(err, idErr) {
		t.Fatalf("expected idErr, got: %v", err)
	}
}

func Test_SaveInvitationUsecase_SaveError(t *testing.T) {
	t.Parallel()

	saveErr := errors.New("save failed")
	repo := invitationmocks.NewMockInvitationRepository()
	repo.SaveInvitationFunc = func(ctx context.Context, invitation *domain.Invitation) error {
		return saveErr
	}

	uc := NewSaveInvitationUsecase(repo, invitationmocks.NewMockIDGenerator())

	_, err := uc.Execute(context.Background(), "invitee@example.com", invitationmocks.DefaultInviterID)
	if !errors.Is(err, saveErr) {
		t.Fatalf("expected saveErr, got: %v", err)
	}
}

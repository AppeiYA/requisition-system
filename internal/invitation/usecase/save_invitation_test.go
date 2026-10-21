package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationmocks "github.com/AppeiYA/requisition-system/internal/invitation/mocks"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

func saveCmd() invitationin.SaveInvitationCommand {
	return invitationin.SaveInvitationCommand{
		Email:        "invitee@example.com",
		FirstName:    "Jane",
		LastName:     "Doe",
		Phone:        "+14155551234",
		StaffID:      invitationmocks.DefaultStaffID,
		DepartmentID: invitationmocks.DefaultDepartmentID,
		JobTitle:     invitationmocks.DefaultJobTitle,
		InviterID:    invitationmocks.DefaultInviterID,
	}
}

func Test_SaveInvitationUsecase_Success(t *testing.T) {
	t.Parallel()

	var saved *domain.Invitation
	repo := invitationmocks.NewMockInvitationRepository()
	repo.SaveInvitationFunc = func(ctx context.Context, invitation *domain.Invitation) error {
		saved = invitation
		return nil
	}

	uc := NewSaveInvitationUsecase(repo, invitationmocks.NewMockIDGenerator())

	res, err := uc.Execute(context.Background(), saveCmd())
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
	if saved.FirstName().Value() != "Jane" || saved.LastName().Value() != "Doe" {
		t.Errorf("expected saved name %q %q, got %q %q", "Jane", "Doe", saved.FirstName().Value(), saved.LastName().Value())
	}
	if saved.Phone().Value() != "+14155551234" {
		t.Errorf("expected saved phone %q, got %q", "+14155551234", saved.Phone().Value())
	}
	if saved.StaffID() != invitationmocks.DefaultStaffID {
		t.Errorf("expected saved staff id %q, got %q", invitationmocks.DefaultStaffID, saved.StaffID())
	}
	if saved.DepartmentID() != invitationmocks.DefaultDepartmentID {
		t.Errorf("expected saved department id %q, got %q", invitationmocks.DefaultDepartmentID, saved.DepartmentID())
	}
	if saved.JobTitle() != invitationmocks.DefaultJobTitle {
		t.Errorf("expected saved job title %q, got %q", invitationmocks.DefaultJobTitle, saved.JobTitle())
	}

	remaining := time.Until(saved.ExpiresAt())
	if remaining <= 23*time.Hour || remaining > 24*time.Hour {
		t.Errorf("expected ~24h expiry, got %v", remaining)
	}
}

func Test_SaveInvitationUsecase_InvalidEmail(t *testing.T) {
	t.Parallel()

	uc := NewSaveInvitationUsecase(invitationmocks.NewMockInvitationRepository(), invitationmocks.NewMockIDGenerator())

	cmd := saveCmd()
	cmd.Email = "not-an-email"

	_, err := uc.Execute(context.Background(), cmd)
	if err != shareddomain.ErrInvalidEmailFormat {
		t.Fatalf("expected ErrInvalidEmailFormat, got: %v", err)
	}
}

func Test_SaveInvitationUsecase_InvalidName(t *testing.T) {
	t.Parallel()

	uc := NewSaveInvitationUsecase(invitationmocks.NewMockInvitationRepository(), invitationmocks.NewMockIDGenerator())

	cmd := saveCmd()
	cmd.FirstName = "J"

	_, err := uc.Execute(context.Background(), cmd)
	expected := shareddomain.ErrUsernameTooShort(shareddomain.FirstName)
	if err == nil || err.Error() != expected.Error() {
		t.Fatalf("expected %v, got: %v", expected, err)
	}
}

func Test_SaveInvitationUsecase_InvalidPhone(t *testing.T) {
	t.Parallel()

	uc := NewSaveInvitationUsecase(invitationmocks.NewMockInvitationRepository(), invitationmocks.NewMockIDGenerator())

	cmd := saveCmd()
	cmd.Phone = "not-a-phone"

	_, err := uc.Execute(context.Background(), cmd)
	if err != shareddomain.ErrInvalidPhoneNumber {
		t.Fatalf("expected ErrInvalidPhoneNumber, got: %v", err)
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

	_, err := uc.Execute(context.Background(), saveCmd())
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

	_, err := uc.Execute(context.Background(), saveCmd())
	if !errors.Is(err, saveErr) {
		t.Fatalf("expected saveErr, got: %v", err)
	}
}

func Test_SaveInvitationUsecase_ExpiresStalePendingInvitation(t *testing.T) {
	t.Parallel()

	stale, err := domain.ReconstituteInvitationWithStatus(
		"inv-stale",
		"Jane",
		"Doe",
		"+14155551234",
		"STF-1234",
		"dpt-1234",
		"Software Engineer",
		"invitee@example.com",
		domain.HashInvitationToken("stale-token"),
		"usr-1234",
		domain.InvitationStatusPending.String(),
		time.Now().Add(-time.Hour),
		nil,
		nil,
		time.Now().Add(-25*time.Hour),
	)
	if err != nil {
		t.Fatalf("failed to create stale invitation: %v", err)
	}

	var expiredID string
	var expiredStatus domain.InvitationStatus
	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetInvitationByEmailFunc = func(ctx context.Context, email string) (*domain.Invitation, error) {
		return stale, nil
	}
	repo.UpdateInvitationStatusFunc = func(ctx context.Context, invitationID string, status domain.InvitationStatus) error {
		expiredID = invitationID
		expiredStatus = status
		return nil
	}

	uc := NewSaveInvitationUsecase(repo, invitationmocks.NewMockIDGenerator())
	if _, err := uc.Execute(context.Background(), saveCmd()); err != nil {
		t.Fatalf("expected stale invitation to be replaced, got: %v", err)
	}
	if expiredID != "inv-stale" {
		t.Fatalf("expected stale invitation %q to be expired, got %q", "inv-stale", expiredID)
	}
	if expiredStatus != domain.InvitationStatusExpired {
		t.Fatalf("expected expired status, got %q", expiredStatus)
	}
}

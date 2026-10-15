package adminexternal

import (
	"context"
	"errors"
	"testing"
	"time"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationmocks "github.com/AppeiYA/requisition-system/internal/invitation/mocks"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
)

type stubSaveInvitation struct {
	res *invitationin.SaveInvitationResponse
	err error
	got invitationin.SaveInvitationCommand
}

func (s *stubSaveInvitation) Execute(ctx context.Context, cmd invitationin.SaveInvitationCommand) (*invitationin.SaveInvitationResponse, error) {
	s.got = cmd
	return s.res, s.err
}

func Test_AdminInvitationAdapter_CheckActiveInvitationExists(t *testing.T) {
	t.Parallel()

	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetActiveInvitationByEmailFunc = func(ctx context.Context, email string) (*invitationdomain.Invitation, error) {
		return invitationmocks.ValidInvitation(), nil
	}

	adapter := NewAdminInvitationAdapter(&stubSaveInvitation{}, repo)

	exists, err := adapter.CheckActiveInvitationExists(context.Background(), mustEmail(t, "invitee@example.com"))
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if !exists {
		t.Fatal("expected an active invitation to exist")
	}
}

func Test_AdminInvitationAdapter_CheckActiveInvitationExists_NotFound(t *testing.T) {
	t.Parallel()

	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetActiveInvitationByEmailFunc = func(ctx context.Context, email string) (*invitationdomain.Invitation, error) {
		return nil, invitationdomain.ErrInvitationNotFound
	}

	adapter := NewAdminInvitationAdapter(&stubSaveInvitation{}, repo)

	exists, err := adapter.CheckActiveInvitationExists(context.Background(), mustEmail(t, "invitee@example.com"))
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if exists {
		t.Fatal("expected no active invitation")
	}
}

func Test_AdminInvitationAdapter_CheckActiveInvitationExists_Error(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db down")
	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetActiveInvitationByEmailFunc = func(ctx context.Context, email string) (*invitationdomain.Invitation, error) {
		return nil, dbErr
	}

	adapter := NewAdminInvitationAdapter(&stubSaveInvitation{}, repo)

	_, err := adapter.CheckActiveInvitationExists(context.Background(), mustEmail(t, "invitee@example.com"))
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}

func Test_AdminInvitationAdapter_SaveInvitation(t *testing.T) {
	t.Parallel()

	expiresAt := time.Now().Add(24 * time.Hour)
	save := &stubSaveInvitation{res: &invitationin.SaveInvitationResponse{
		ID:        "inv-1",
		Token:     "token",
		ExpiresAt: expiresAt,
	}}

	adapter := NewAdminInvitationAdapter(save, invitationmocks.NewMockInvitationRepository())

	res, err := adapter.SaveInvitation(context.Background(), adminout.SaveInvitationParams{
		Email:        mustEmail(t, "invitee@example.com"),
		FirstName:    "Jane",
		LastName:     "Doe",
		Phone:        "+14155551234",
		StaffID:      "STF-1234",
		DepartmentID: "dpt-1",
		JobTitle:     "Engineer",
		InviterID:    "usr-1",
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if res.ID != "inv-1" || res.Token != "token" {
		t.Errorf("unexpected result: %+v", res)
	}
	if !res.ExpiresAt.Equal(expiresAt) {
		t.Errorf("expected expiry %v, got %v", expiresAt, res.ExpiresAt)
	}
	if save.got.Email != "invitee@example.com" {
		t.Errorf("expected email to be forwarded, got %q", save.got.Email)
	}
	if save.got.FirstName != "Jane" || save.got.LastName != "Doe" || save.got.Phone != "+14155551234" {
		t.Errorf("expected prefill to be forwarded, got %+v", save.got)
	}
	if save.got.StaffID != "STF-1234" {
		t.Errorf("expected staff id to be forwarded, got %q", save.got.StaffID)
	}
	if save.got.DepartmentID != "dpt-1" {
		t.Errorf("expected department id to be forwarded, got %q", save.got.DepartmentID)
	}
	if save.got.JobTitle != "Engineer" {
		t.Errorf("expected job title to be forwarded, got %q", save.got.JobTitle)
	}
}

func Test_AdminInvitationAdapter_SaveInvitation_Error(t *testing.T) {
	t.Parallel()

	saveErr := errors.New("save failed")
	adapter := NewAdminInvitationAdapter(&stubSaveInvitation{err: saveErr}, invitationmocks.NewMockInvitationRepository())

	_, err := adapter.SaveInvitation(context.Background(), adminout.SaveInvitationParams{
		Email:     mustEmail(t, "invitee@example.com"),
		InviterID: "usr-1",
	})
	if !errors.Is(err, saveErr) {
		t.Fatalf("expected saveErr, got: %v", err)
	}
}

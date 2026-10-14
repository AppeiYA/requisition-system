package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationmocks "github.com/AppeiYA/requisition-system/internal/invitation/mocks"
)

func Test_ValidateInvitationUsecase_Valid(t *testing.T) {
	t.Parallel()

	var gotHash string
	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
		gotHash = tokenHash
		return pendingInvitation(), nil
	}

	uc := NewValidateInvitationUsecase(repo)

	res, err := uc.Execute(context.Background(), testInvitationToken)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if !res.Valid {
		t.Fatal("expected invitation to be valid")
	}
	if res.Email != testInvitationEmail {
		t.Errorf("expected email %q, got %q", testInvitationEmail, res.Email)
	}
	if gotHash != domain.HashInvitationToken(testInvitationToken) {
		t.Errorf("lookup used wrong token hash")
	}
}

func Test_ValidateInvitationUsecase_EmptyToken(t *testing.T) {
	t.Parallel()

	uc := NewValidateInvitationUsecase(invitationmocks.NewMockInvitationRepository())

	res, err := uc.Execute(context.Background(), "")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if res.Valid {
		t.Fatal("expected invitation to be invalid")
	}
}

func Test_ValidateInvitationUsecase_NotFound(t *testing.T) {
	t.Parallel()

	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
		return nil, domain.ErrInvitationNotFound
	}

	uc := NewValidateInvitationUsecase(repo)

	res, err := uc.Execute(context.Background(), testInvitationToken)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if res.Valid {
		t.Fatal("expected invitation to be invalid")
	}
}

func Test_ValidateInvitationUsecase_InvalidStates(t *testing.T) {
	t.Parallel()

	past := time.Now().Add(-time.Hour)
	present := time.Now()

	tests := []struct {
		name string
		inv  *domain.Invitation
	}{
		{name: "expired", inv: invitationWithState(past, nil, nil)},
		{name: "revoked", inv: invitationWithState(time.Now().Add(time.Hour), nil, &present)},
		{name: "accepted", inv: invitationWithState(time.Now().Add(time.Hour), &present, nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := invitationmocks.NewMockInvitationRepository()
			repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
				return tt.inv, nil
			}

			uc := NewValidateInvitationUsecase(repo)

			res, err := uc.Execute(context.Background(), testInvitationToken)
			if err != nil {
				t.Fatalf("expected nil error, got: %v", err)
			}
			if res.Valid {
				t.Fatal("expected invitation to be invalid")
			}
			if res.Email != testInvitationEmail {
				t.Errorf("expected email %q, got %q", testInvitationEmail, res.Email)
			}
		})
	}
}

func Test_ValidateInvitationUsecase_RepoError(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db connection failure")
	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
		return nil, dbErr
	}

	uc := NewValidateInvitationUsecase(repo)

	_, err := uc.Execute(context.Background(), testInvitationToken)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}

func Test_ValidateInvitationUsecase_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	uc := NewValidateInvitationUsecase(invitationmocks.NewMockInvitationRepository())

	_, err := uc.Execute(ctx, testInvitationToken)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

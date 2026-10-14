package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identitymocks "github.com/AppeiYA/requisition-system/internal/identity/mocks"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

func Test_GetMeUsecase_Success(t *testing.T) {
	t.Parallel()

	user := identitymocks.ValidUser()

	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByIDFunc = func(ctx context.Context, id string) (*domain.User, error) {
		if id != identitymocks.DefaultID {
			t.Errorf("expected user ID %q, got %q", identitymocks.DefaultID, id)
		}
		return user, nil
	}

	uc := NewGetMeUsecase(repo)

	response, err := uc.Execute(context.Background(), identitymocks.DefaultID)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if response == nil {
		t.Fatal("expected response, got nil")
	}

	if response.ID != user.ID() {
		t.Errorf("expected id %q, got %q", user.ID(), response.ID)
	}
	if response.Email != user.Email().Value() {
		t.Errorf("expected email %q, got %q", user.Email().Value(), response.Email)
	}
	if response.FirstName != user.FirstName().Value() {
		t.Errorf("expected first name %q, got %q", user.FirstName().Value(), response.FirstName)
	}
	if response.LastName != user.LastName().Value() {
		t.Errorf("expected last name %q, got %q", user.LastName().Value(), response.LastName)
	}
	if !response.CreatedAt.Equal(user.CreatedAt()) {
		t.Errorf("expected createdAt %v, got %v", user.CreatedAt(), response.CreatedAt)
	}
	if response.UpdatedAt != nil {
		t.Errorf("expected UpdatedAt nil, got %v", *response.UpdatedAt)
	}
}

func Test_GetMeUsecase_GetUserByID_RepoError(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db connection failure")

	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByIDFunc = func(ctx context.Context, id string) (*domain.User, error) {
		return nil, dbErr
	}

	uc := NewGetMeUsecase(repo)

	response, err := uc.Execute(context.Background(), identitymocks.DefaultID)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
	if response != nil {
		t.Fatalf("expected nil response on error, got: %v", response)
	}
}

func Test_GetMeUsecase_GetUserByID_NotFound(t *testing.T) {
	t.Parallel()

	notFoundErr := customerrors.NewNotFoundError("user not found")

	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByIDFunc = func(ctx context.Context, id string) (*domain.User, error) {
		return nil, notFoundErr
	}

	uc := NewGetMeUsecase(repo)

	_, err := uc.Execute(context.Background(), identitymocks.DefaultID)
	if err == nil {
		t.Fatal("expected not found error, got nil")
	}
	if err.Error() != notFoundErr.Error() {
		t.Fatalf("expected %v, got: %v", notFoundErr, err)
	}
}

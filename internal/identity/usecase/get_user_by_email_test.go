package usecase

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identitymocks "github.com/AppeiYA/requisition-system/internal/identity/mocks"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

func Test_GetUserByEmailUsecase_Success(t *testing.T) {
	t.Parallel()

	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *shareddomain.Email) (*domain.User, error) {
		return identitymocks.ValidUser(), nil
	}

	uc := NewGetUserByEmailUsecase(repo)

	res, err := uc.Execute(context.Background(), identitymocks.DefaultEmail)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if res.ID != identitymocks.DefaultID {
		t.Errorf("expected ID %q, got %q", identitymocks.DefaultID, res.ID)
	}
	if res.Email != identitymocks.DefaultEmail {
		t.Errorf("expected email %q, got %q", identitymocks.DefaultEmail, res.Email)
	}
	if res.FirstName != identitymocks.DefaultFirstName {
		t.Errorf("expected first name %q, got %q", identitymocks.DefaultFirstName, res.FirstName)
	}
	if res.LastName != identitymocks.DefaultLastName {
		t.Errorf("expected last name %q, got %q", identitymocks.DefaultLastName, res.LastName)
	}
}

func Test_GetUserByEmailUsecase_InvalidEmail(t *testing.T) {
	t.Parallel()

	uc := NewGetUserByEmailUsecase(identitymocks.NewMockUsersRepository())

	_, err := uc.Execute(context.Background(), "not-an-email")
	if err != shareddomain.ErrInvalidEmailFormat {
		t.Fatalf("expected ErrInvalidEmailFormat, got: %v", err)
	}
}

func Test_GetUserByEmailUsecase_UserNotFound(t *testing.T) {
	t.Parallel()

	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *shareddomain.Email) (*domain.User, error) {
		return nil, customerrors.NewNotFoundError("user not found")
	}

	uc := NewGetUserByEmailUsecase(repo)

	_, err := uc.Execute(context.Background(), identitymocks.DefaultEmail)
	var custErr *customerrors.CustomError
	if !errors.As(err, &custErr) || custErr.Code != http.StatusNotFound {
		t.Fatalf("expected not found custom error, got: %v", err)
	}
}

func Test_GetUserByEmailUsecase_NilUser(t *testing.T) {
	t.Parallel()

	repo := identitymocks.NewMockUsersRepository()

	uc := NewGetUserByEmailUsecase(repo)

	_, err := uc.Execute(context.Background(), identitymocks.DefaultEmail)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
}

func Test_GetUserByEmailUsecase_RepoError(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db connection failure")
	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *shareddomain.Email) (*domain.User, error) {
		return nil, dbErr
	}

	uc := NewGetUserByEmailUsecase(repo)

	_, err := uc.Execute(context.Background(), identitymocks.DefaultEmail)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}

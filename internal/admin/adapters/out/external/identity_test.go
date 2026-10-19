package adminexternal

import (
	"context"
	"errors"
	"testing"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

type stubGetUserByEmail struct {
	user *identityin.UserResponse
	err  error
}

func (s *stubGetUserByEmail) Execute(ctx context.Context, email string) (*identityin.UserResponse, error) {
	return s.user, s.err
}

func mustEmail(t *testing.T, value string) *shareddomain.Email {
	t.Helper()
	email, err := shareddomain.NewEmail(value)
	if err != nil {
		t.Fatalf("invalid test email %q: %v", value, err)
	}
	return email
}

func Test_AdminIdentityAdapter_CheckUserExists_True(t *testing.T) {
	t.Parallel()

	adapter := NewAdminIdentityAdapter(&stubGetUserByEmail{user: &identityin.UserResponse{ID: "usr-1"}})

	exists, err := adapter.CheckUserExists(context.Background(), mustEmail(t, "user@example.com"))
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if !exists {
		t.Fatal("expected user to exist")
	}
}

func Test_AdminIdentityAdapter_CheckUserExists_NotFound(t *testing.T) {
	t.Parallel()

	adapter := NewAdminIdentityAdapter(&stubGetUserByEmail{err: customerrors.NewNotFoundError("user not found")})

	exists, err := adapter.CheckUserExists(context.Background(), mustEmail(t, "user@example.com"))
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if exists {
		t.Fatal("expected user not to exist")
	}
}

func Test_AdminIdentityAdapter_CheckUserExists_NilUser(t *testing.T) {
	t.Parallel()

	adapter := NewAdminIdentityAdapter(&stubGetUserByEmail{})

	exists, err := adapter.CheckUserExists(context.Background(), mustEmail(t, "user@example.com"))
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if exists {
		t.Fatal("expected user not to exist")
	}
}

func Test_AdminIdentityAdapter_CheckUserExists_Error(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db down")
	adapter := NewAdminIdentityAdapter(&stubGetUserByEmail{err: dbErr})

	_, err := adapter.CheckUserExists(context.Background(), mustEmail(t, "user@example.com"))
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}

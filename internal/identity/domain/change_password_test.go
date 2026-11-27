package domain

import (
	"errors"
	"testing"

	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

func Test_User_ChangePassword(t *testing.T) {
	t.Parallel()

	email, _ := shareddomain.NewEmail("change.password@example.com")
	oldHash, _ := NewPasswordHash("old_hashed_password_123")
	user, err := NewUser("usr-change", "Change", "Password", "", "", email, oldHash)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	newHash, err := NewPasswordHash("new_hashed_password_123")
	if err != nil {
		t.Fatalf("new password hash: %v", err)
	}
	if err := user.ChangePassword(newHash); err != nil {
		t.Fatalf("change password: %v", err)
	}
	if user.PasswordHash() != newHash || user.PasswordHash() == oldHash {
		t.Fatalf("password hash was not replaced")
	}
	if user.UpdatedAt() == nil {
		t.Fatalf("updatedAt was not recorded")
	}

	if err := user.ChangePassword(nil); !errors.Is(err, ErrInvalidPasswordHash) {
		t.Fatalf("nil hash error = %v, want %v", err, ErrInvalidPasswordHash)
	}
}

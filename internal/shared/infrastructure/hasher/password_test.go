package hasher

import (
	"strings"
	"testing"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"golang.org/x/crypto/bcrypt"
)

var _ identityout.PasswordHasher = (*PasswordHasher)(nil)

func Test_PasswordHasher_HashPassword_Success(t *testing.T) {
	hasher := NewPasswordHasher()
	password := "securePassword123!"

	hash, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty hash string")
	}

	if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
		t.Errorf("expected bcrypt hash prefix ($2a$ or $2b$), got: %s", hash)
	}
}

func Test_PasswordHasher_ComparePassword_Success(t *testing.T) {
	hasher := NewPasswordHasher()
	password := "mySecretPassword"

	hash, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	err = hasher.ComparePassword(hash, password)
	if err != nil {
		t.Errorf("expected valid password comparison, got error: %v", err)
	}
}

func Test_PasswordHasher_ComparePassword_Mismatch(t *testing.T) {
	hasher := NewPasswordHasher()
	password := "correctPassword"
	wrongPassword := "incorrectPassword"

	hash, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	err = hasher.ComparePassword(hash, wrongPassword)
	if err == nil {
		t.Error("expected error for mismatched password, got nil")
	}
	if err != bcrypt.ErrMismatchedHashAndPassword {
		t.Errorf("expected bcrypt.ErrMismatchedHashAndPassword, got: %v", err)
	}
}

func Test_PasswordHasher_ComparePassword_InvalidHash(t *testing.T) {
	hasher := NewPasswordHasher()

	err := hasher.ComparePassword("not-a-valid-bcrypt-hash", "password")
	if err == nil {
		t.Error("expected error when comparing with invalid hash format, got nil")
	}
}

func Test_PasswordHasher_PasswordTooLong(t *testing.T) {
	hasher := NewPasswordHasher()
	tooLongPassword := strings.Repeat("a", 73) // bcrypt maximum password length is 72 bytes

	_, err := hasher.HashPassword(tooLongPassword)
	if err == nil {
		t.Error("expected error when hashing password exceeding 72 bytes, got nil")
	}
	if err != bcrypt.ErrPasswordTooLong {
		t.Errorf("expected bcrypt.ErrPasswordTooLong, got: %v", err)
	}
}

func Test_PasswordHasher_ZeroCostDefaults(t *testing.T) {
	// struct initialized with zero cost should default to bcrypt.DefaultCost
	hasher := &PasswordHasher{}
	password := "testDefaultCost"

	hash, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("unexpected error getting hash cost: %v", err)
	}

	if cost != bcrypt.DefaultCost {
		t.Errorf("expected cost %d, got %d", bcrypt.DefaultCost, cost)
	}
}

func Test_PasswordHasher_CustomCost(t *testing.T) {
	hasher := NewPasswordHasherWithCost(bcrypt.MinCost)
	password := "testMinCost"

	hash, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("unexpected error getting hash cost: %v", err)
	}

	if cost != bcrypt.MinCost {
		t.Errorf("expected cost %d, got %d", bcrypt.MinCost, cost)
	}
}

func Test_PasswordHasher_Aliases(t *testing.T) {
	hasher := NewPasswordHasher()
	password := "aliasTestPassword"

	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("unexpected error from Hash alias: %v", err)
	}

	err = hasher.Compare(hash, password)
	if err != nil {
		t.Errorf("unexpected error from Compare alias: %v", err)
	}
}

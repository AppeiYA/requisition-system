package mock_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/identity/adapters/out/mock"
	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

var (
	_ identityout.UsersRepository = (*mock.MockUsersRepository)(nil)
	_ identityout.PasswordHasher  = (*mock.MockPasswordHasher)(nil)
	_ identityout.IDGenerator     = (*mock.MockIDGenerator)(nil)
)

func Test_MockUsersRepository(t *testing.T) {
	t.Parallel()

	repo := mock.NewMockUsersRepository()
	ctx := context.Background()
	user := mock.ValidUser()
	email := mock.ValidEmail()

	// Default behavior returns nil / no error
	if err := repo.SaveUser(ctx, user); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if u, err := repo.GetUserByEmail(ctx, email); u != nil || err != nil {
		t.Fatalf("expected nil user and nil error, got u=%v, err=%v", u, err)
	}
	if u, err := repo.GetUserByID(ctx, "usr-1"); u != nil || err != nil {
		t.Fatalf("expected nil user and nil error, got u=%v, err=%v", u, err)
	}
	if err := repo.UpdateUserStatus(ctx, user); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := repo.UpdateUserPassword(ctx, user); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := repo.UpdateUser(ctx, user); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Function overrides
	testErr := errors.New("custom error")
	repo.SaveUserFunc = func(ctx context.Context, u *domain.User) error { return testErr }
	if err := repo.SaveUser(ctx, user); !errors.Is(err, testErr) {
		t.Fatalf("expected custom error, got %v", err)
	}

	repo.GetUserByEmailFunc = func(ctx context.Context, e *domain.Email) (*domain.User, error) {
		return user, nil
	}
	if found, err := repo.GetUserByEmail(ctx, email); found != user || err != nil {
		t.Fatalf("expected user, got %v, err=%v", found, err)
	}
}

func Test_MockPasswordHasher(t *testing.T) {
	t.Parallel()

	hasher := mock.NewMockPasswordHasher()

	// Default behavior
	hash, err := hasher.HashPassword("mySecret123!")
	if err != nil || hash != "hashed_mySecret123!" {
		t.Fatalf("expected hashed_mySecret123!, got %v, err=%v", hash, err)
	}

	if err := hasher.ComparePassword(hash, "mySecret123!"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Override
	testErr := errors.New("hasher error")
	hasher.HashPasswordFunc = func(password string) (string, error) {
		return "", testErr
	}
	if _, err := hasher.HashPassword("pwd"); !errors.Is(err, testErr) {
		t.Fatalf("expected custom error, got %v", err)
	}
}

func Test_MockIDGenerator(t *testing.T) {
	t.Parallel()

	idGen := mock.NewMockIDGenerator()

	// Default behavior
	id, err := idGen.GenerateID("usr")
	if err != nil || id != "usr-12345678" {
		t.Fatalf("expected usr-12345678, got %s, err=%v", id, err)
	}

	// Override
	testErr := errors.New("id generator error")
	idGen.GenerateIDFunc = func(prefix string) (string, error) {
		return "", testErr
	}
	if _, err := idGen.GenerateID("usr"); !errors.Is(err, testErr) {
		t.Fatalf("expected custom error, got %v", err)
	}
}

func Test_Fixtures_UserBuilder(t *testing.T) {
	t.Parallel()

	user := mock.ValidUser()
	if user == nil {
		t.Fatal("expected valid user")
	}

	now := time.Now()
	customUser := mock.NewUserBuilder().
		WithID("usr-custom").
		WithFirstName("Alice").
		WithLastName("Smith").
		WithEmail("alice.smith@example.com").
		WithPasswordHash("customhash").
		WithStatus(string(domain.ACTIVE)).
		WithCreatedAt(now).
		WithUpdatedAt(&now).
		MustBuild()

	if customUser.FirstName().Value() != "Alice" {
		t.Errorf("expected 'Alice', got %s", customUser.FirstName().Value())
	}
	if customUser.LastName().Value() != "Smith" {
		t.Errorf("expected 'Smith', got %s", customUser.LastName().Value())
	}
	if customUser.Email().Value() != "alice.smith@example.com" {
		t.Errorf("expected 'alice.smith@example.com', got %s", customUser.Email().Value())
	}
}


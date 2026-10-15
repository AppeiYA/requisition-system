package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identitymocks "github.com/AppeiYA/requisition-system/internal/identity/mocks"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

func Test_RegisterUserUsecase_Success(t *testing.T) {
	t.Parallel()

	var savedUser *domain.User

	repo := identitymocks.NewMockUsersRepository()
	repo.SaveUserFunc = func(ctx context.Context, user *domain.User) error {
		savedUser = user
		return nil
	}
	hasher := identitymocks.NewMockPasswordHasher()
	idGen := identitymocks.NewMockIDGenerator()

	uc := NewRegisterUserUsecase(repo, hasher, idGen)

	dto := identityin.RegisterUserDTO{
		FirstName: identitymocks.DefaultFirstName,
		LastName:  identitymocks.DefaultLastName,
		Phone:     identitymocks.DefaultPhone,
		StaffID:   identitymocks.DefaultStaffID,
		Email:     identitymocks.DefaultEmail,
		Password:  identitymocks.DefaultPassword,
	}

	gotID, err := uc.Execute(context.Background(), dto)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if savedUser == nil {
		t.Fatal("expected user to be saved, got nil")
	}

	if gotID == "" || gotID != savedUser.ID() {
		t.Errorf("expected returned id %q to match saved user id", gotID)
	}

	if savedUser.FirstName().Value() != identitymocks.DefaultFirstName {
		t.Errorf("expected first name '%s', got '%s'", identitymocks.DefaultFirstName, savedUser.FirstName().Value())
	}
	if savedUser.LastName().Value() != identitymocks.DefaultLastName {
		t.Errorf("expected last name '%s', got '%s'", identitymocks.DefaultLastName, savedUser.LastName().Value())
	}
	if savedUser.Phone().Value() != identitymocks.DefaultPhone {
		t.Errorf("expected phone '%s', got '%s'", identitymocks.DefaultPhone, savedUser.Phone().Value())
	}
	if savedUser.StaffID() != identitymocks.DefaultStaffID {
		t.Errorf("expected staff id '%s', got '%s'", identitymocks.DefaultStaffID, savedUser.StaffID())
	}
	if savedUser.Email().Value() != identitymocks.DefaultEmail {
		t.Errorf("expected email '%s', got '%s'", identitymocks.DefaultEmail, savedUser.Email().Value())
	}
	if savedUser.Status() != domain.ACTIVE {
		t.Errorf("expected status 'active', got '%s'", savedUser.Status())
	}
}

func Test_RegisterUserUsecase_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	uc := NewRegisterUserUsecase(identitymocks.NewMockUsersRepository(), identitymocks.NewMockPasswordHasher(), identitymocks.NewMockIDGenerator())

	dto := identityin.RegisterUserDTO{
		FirstName: identitymocks.DefaultFirstName,
		LastName:  identitymocks.DefaultLastName,
		Email:     identitymocks.DefaultEmail,
		Password:  identitymocks.DefaultPassword,
	}

	_, err := uc.Execute(ctx, dto)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func Test_RegisterUserUsecase_PasswordValidation(t *testing.T) {
	t.Parallel()

	uc := NewRegisterUserUsecase(identitymocks.NewMockUsersRepository(), identitymocks.NewMockPasswordHasher(), identitymocks.NewMockIDGenerator())

	tests := []struct {
		name        string
		password    string
		expectedErr error
	}{
		{
			name:        "empty password",
			password:    "",
			expectedErr: domain.ErrPasswordIsRequired,
		},
		{
			name:        "whitespace only",
			password:    "   ",
			expectedErr: domain.ErrInvalidPassword,
		},
		{
			name:        "too short password",
			password:    "Pass1!",
			expectedErr: domain.ErrInvalidPassword,
		},
		{
			name:        "missing uppercase",
			password:    "password123!",
			expectedErr: domain.ErrInvalidPassword,
		},
		{
			name:        "missing lowercase",
			password:    "PASSWORD123!",
			expectedErr: domain.ErrInvalidPassword,
		},
		{
			name:        "missing digit",
			password:    "Password!!!!!",
			expectedErr: domain.ErrInvalidPassword,
		},
		{
			name:        "missing special character",
			password:    "Password12345",
			expectedErr: domain.ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dto := identityin.RegisterUserDTO{
				FirstName: identitymocks.DefaultFirstName,
				LastName:  identitymocks.DefaultLastName,
				Email:     identitymocks.DefaultEmail,
				Password:  tt.password,
			}

			_, err := uc.Execute(context.Background(), dto)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != tt.expectedErr.Error() {
				t.Fatalf("expected %v, got: %v", tt.expectedErr, err)
			}
		})
	}
}

func Test_RegisterUserUsecase_InvalidEmail(t *testing.T) {
	t.Parallel()

	uc := NewRegisterUserUsecase(identitymocks.NewMockUsersRepository(), identitymocks.NewMockPasswordHasher(), identitymocks.NewMockIDGenerator())

	dto := identityin.RegisterUserDTO{
		FirstName: identitymocks.DefaultFirstName,
		LastName:  identitymocks.DefaultLastName,
		Email:     "invalid-email",
		Password:  identitymocks.DefaultPassword,
	}

	_, err := uc.Execute(context.Background(), dto)
	if err != shareddomain.ErrInvalidEmailFormat {
		t.Fatalf("expected ErrInvalidEmailFormat, got: %v", err)
	}
}

func Test_RegisterUserUsecase_UserAlreadyExists(t *testing.T) {
	t.Parallel()

	existingUser := identitymocks.ValidUser()
	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *shareddomain.Email) (*domain.User, error) {
		return existingUser, nil
	}

	uc := NewRegisterUserUsecase(repo, identitymocks.NewMockPasswordHasher(), identitymocks.NewMockIDGenerator())

	dto := identityin.RegisterUserDTO{
		FirstName: identitymocks.DefaultFirstName,
		LastName:  identitymocks.DefaultLastName,
		Email:     identitymocks.DefaultEmail,
		Password:  identitymocks.DefaultPassword,
	}

	_, err := uc.Execute(context.Background(), dto)
	if err == nil {
		t.Fatal("expected ErrUserAlreadyExists, got nil")
	}
	if err != ErrUserAlreadyExists && err.Error() != ErrUserAlreadyExists.Error() {
		t.Fatalf("expected ErrUserAlreadyExists, got: %v", err)
	}
}

func Test_RegisterUserUsecase_GetUserByEmail_NotFoundErrorIgnored(t *testing.T) {
	t.Parallel()

	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *shareddomain.Email) (*domain.User, error) {
		return nil, customerrors.NewNotFoundError("user not found")
	}

	uc := NewRegisterUserUsecase(repo, identitymocks.NewMockPasswordHasher(), identitymocks.NewMockIDGenerator())

	dto := identityin.RegisterUserDTO{
		FirstName: identitymocks.DefaultFirstName,
		LastName:  identitymocks.DefaultLastName,
		Email:     identitymocks.DefaultEmail,
		Password:  identitymocks.DefaultPassword,
	}

	_, err := uc.Execute(context.Background(), dto)
	if err != nil {
		t.Fatalf("expected nil error when repo returns NotFound, got: %v", err)
	}
}

func Test_RegisterUserUsecase_GetUserByEmail_RepoError(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db connection failure")
	repo := identitymocks.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *shareddomain.Email) (*domain.User, error) {
		return nil, dbErr
	}

	uc := NewRegisterUserUsecase(repo, identitymocks.NewMockPasswordHasher(), identitymocks.NewMockIDGenerator())

	dto := identityin.RegisterUserDTO{
		FirstName: identitymocks.DefaultFirstName,
		LastName:  identitymocks.DefaultLastName,
		Email:     identitymocks.DefaultEmail,
		Password:  identitymocks.DefaultPassword,
	}

	_, err := uc.Execute(context.Background(), dto)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}

func Test_RegisterUserUsecase_IDGeneratorError(t *testing.T) {
	t.Parallel()

	idGenErr := errors.New("id generation failed")
	idGen := identitymocks.NewMockIDGenerator()
	idGen.GenerateIDFunc = func(prefix string) (string, error) {
		return "", idGenErr
	}

	uc := NewRegisterUserUsecase(identitymocks.NewMockUsersRepository(), identitymocks.NewMockPasswordHasher(), idGen)

	dto := identityin.RegisterUserDTO{
		FirstName: identitymocks.DefaultFirstName,
		LastName:  identitymocks.DefaultLastName,
		Email:     identitymocks.DefaultEmail,
		Password:  identitymocks.DefaultPassword,
	}

	_, err := uc.Execute(context.Background(), dto)
	if !errors.Is(err, idGenErr) {
		t.Fatalf("expected idGenErr, got: %v", err)
	}
}

func Test_RegisterUserUsecase_PasswordHasherError(t *testing.T) {
	t.Parallel()

	hasherErr := errors.New("hashing failed")
	hasher := identitymocks.NewMockPasswordHasher()
	hasher.HashPasswordFunc = func(password string) (string, error) {
		return "", hasherErr
	}

	uc := NewRegisterUserUsecase(identitymocks.NewMockUsersRepository(), hasher, identitymocks.NewMockIDGenerator())

	dto := identityin.RegisterUserDTO{
		FirstName: identitymocks.DefaultFirstName,
		LastName:  identitymocks.DefaultLastName,
		Email:     identitymocks.DefaultEmail,
		Password:  identitymocks.DefaultPassword,
	}

	_, err := uc.Execute(context.Background(), dto)
	if !errors.Is(err, hasherErr) {
		t.Fatalf("expected hasherErr, got: %v", err)
	}
}

func Test_RegisterUserUsecase_DomainValidationFailure(t *testing.T) {
	t.Parallel()

	uc := NewRegisterUserUsecase(identitymocks.NewMockUsersRepository(), identitymocks.NewMockPasswordHasher(), identitymocks.NewMockIDGenerator())

	tests := []struct {
		name        string
		dto         identityin.RegisterUserDTO
		expectedErr error
	}{
		{
			name: "short first name",
			dto: identityin.RegisterUserDTO{
				FirstName: "J",
				LastName:  identitymocks.DefaultLastName,
				Email:     identitymocks.DefaultEmail,
				Password:  identitymocks.DefaultPassword,
			},
			expectedErr: domain.ErrUsernameTooShort(domain.FirstName),
		},
		{
			name: "short last name",
			dto: identityin.RegisterUserDTO{
				FirstName: identitymocks.DefaultFirstName,
				LastName:  "D",
				Email:     identitymocks.DefaultEmail,
				Password:  identitymocks.DefaultPassword,
			},
			expectedErr: domain.ErrUsernameTooShort(domain.LastName),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := uc.Execute(context.Background(), tt.dto)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != tt.expectedErr.Error() {
				t.Fatalf("expected %v, got: %v", tt.expectedErr, err)
			}
		})
	}
}

func Test_RegisterUserUsecase_SaveUserError(t *testing.T) {
	t.Parallel()

	saveErr := errors.New("save user failed")
	repo := identitymocks.NewMockUsersRepository()
	repo.SaveUserFunc = func(ctx context.Context, user *domain.User) error {
		return saveErr
	}

	uc := NewRegisterUserUsecase(repo, identitymocks.NewMockPasswordHasher(), identitymocks.NewMockIDGenerator())

	dto := identityin.RegisterUserDTO{
		FirstName: identitymocks.DefaultFirstName,
		LastName:  identitymocks.DefaultLastName,
		Email:     identitymocks.DefaultEmail,
		Password:  identitymocks.DefaultPassword,
	}

	_, err := uc.Execute(context.Background(), dto)
	if !errors.Is(err, saveErr) {
		t.Fatalf("expected saveErr, got: %v", err)
	}
}

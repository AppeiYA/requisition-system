package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/identity/adapters/out/mock"
	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/errors"
)

func Test_RegisterUserUsecase_Success(t *testing.T) {
	t.Parallel()

	var savedUser *domain.User

	repo := mock.NewMockUsersRepository()
	repo.SaveUserFunc = func(ctx context.Context, user *domain.User) error {
		savedUser = user
		return nil
	}
	hasher := mock.NewMockPasswordHasher()
	idGen := mock.NewMockIDGenerator()

	uc := NewRegisterUserUsecase(repo, hasher, idGen)

	dto := RegisterUserDTO{
		FirstName: mock.DefaultFirstName,
		LastName:  mock.DefaultLastName,
		Email:     mock.DefaultEmail,
		Password:  mock.DefaultPassword,
	}

	err := uc.Execute(context.Background(), dto)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if savedUser == nil {
		t.Fatal("expected user to be saved, got nil")
	}

	if savedUser.FirstName().Value() != mock.DefaultFirstName {
		t.Errorf("expected first name '%s', got '%s'", mock.DefaultFirstName, savedUser.FirstName().Value())
	}
	if savedUser.LastName().Value() != mock.DefaultLastName {
		t.Errorf("expected last name '%s', got '%s'", mock.DefaultLastName, savedUser.LastName().Value())
	}
	if savedUser.Email().Value() != mock.DefaultEmail {
		t.Errorf("expected email '%s', got '%s'", mock.DefaultEmail, savedUser.Email().Value())
	}
	if savedUser.Status() != domain.ACTIVE {
		t.Errorf("expected status 'active', got '%s'", savedUser.Status())
	}
}

func Test_RegisterUserUsecase_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	uc := NewRegisterUserUsecase(mock.NewMockUsersRepository(), mock.NewMockPasswordHasher(), mock.NewMockIDGenerator())

	dto := RegisterUserDTO{
		FirstName: mock.DefaultFirstName,
		LastName:  mock.DefaultLastName,
		Email:     mock.DefaultEmail,
		Password:  mock.DefaultPassword,
	}

	err := uc.Execute(ctx, dto)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func Test_RegisterUserUsecase_PasswordValidation(t *testing.T) {
	t.Parallel()

	uc := NewRegisterUserUsecase(mock.NewMockUsersRepository(), mock.NewMockPasswordHasher(), mock.NewMockIDGenerator())

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

			dto := RegisterUserDTO{
				FirstName: mock.DefaultFirstName,
				LastName:  mock.DefaultLastName,
				Email:     mock.DefaultEmail,
				Password:  tt.password,
			}

			err := uc.Execute(context.Background(), dto)
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

	uc := NewRegisterUserUsecase(mock.NewMockUsersRepository(), mock.NewMockPasswordHasher(), mock.NewMockIDGenerator())

	dto := RegisterUserDTO{
		FirstName: mock.DefaultFirstName,
		LastName:  mock.DefaultLastName,
		Email:     "invalid-email",
		Password:  mock.DefaultPassword,
	}

	err := uc.Execute(context.Background(), dto)
	if err != domain.ErrInvalidEmailFormat {
		t.Fatalf("expected ErrInvalidEmailFormat, got: %v", err)
	}
}

func Test_RegisterUserUsecase_UserAlreadyExists(t *testing.T) {
	t.Parallel()

	existingUser := mock.ValidUser()
	repo := mock.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *domain.Email) (*domain.User, error) {
		return existingUser, nil
	}

	uc := NewRegisterUserUsecase(repo, mock.NewMockPasswordHasher(), mock.NewMockIDGenerator())

	dto := RegisterUserDTO{
		FirstName: mock.DefaultFirstName,
		LastName:  mock.DefaultLastName,
		Email:     mock.DefaultEmail,
		Password:  mock.DefaultPassword,
	}

	err := uc.Execute(context.Background(), dto)
	if err == nil {
		t.Fatal("expected ErrUserAlreadyExists, got nil")
	}
	if err != ErrUserAlreadyExists && err.Error() != ErrUserAlreadyExists.Error() {
		t.Fatalf("expected ErrUserAlreadyExists, got: %v", err)
	}
}

func Test_RegisterUserUsecase_GetUserByEmail_NotFoundErrorIgnored(t *testing.T) {
	t.Parallel()

	repo := mock.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *domain.Email) (*domain.User, error) {
		return nil, customerrors.NewNotFoundError("user not found")
	}

	uc := NewRegisterUserUsecase(repo, mock.NewMockPasswordHasher(), mock.NewMockIDGenerator())

	dto := RegisterUserDTO{
		FirstName: mock.DefaultFirstName,
		LastName:  mock.DefaultLastName,
		Email:     mock.DefaultEmail,
		Password:  mock.DefaultPassword,
	}

	err := uc.Execute(context.Background(), dto)
	if err != nil {
		t.Fatalf("expected nil error when repo returns NotFound, got: %v", err)
	}
}

func Test_RegisterUserUsecase_GetUserByEmail_RepoError(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db connection failure")
	repo := mock.NewMockUsersRepository()
	repo.GetUserByEmailFunc = func(ctx context.Context, email *domain.Email) (*domain.User, error) {
		return nil, dbErr
	}

	uc := NewRegisterUserUsecase(repo, mock.NewMockPasswordHasher(), mock.NewMockIDGenerator())

	dto := RegisterUserDTO{
		FirstName: mock.DefaultFirstName,
		LastName:  mock.DefaultLastName,
		Email:     mock.DefaultEmail,
		Password:  mock.DefaultPassword,
	}

	err := uc.Execute(context.Background(), dto)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}

func Test_RegisterUserUsecase_IDGeneratorError(t *testing.T) {
	t.Parallel()

	idGenErr := errors.New("id generation failed")
	idGen := mock.NewMockIDGenerator()
	idGen.GenerateIDFunc = func(prefix string) (string, error) {
		return "", idGenErr
	}

	uc := NewRegisterUserUsecase(mock.NewMockUsersRepository(), mock.NewMockPasswordHasher(), idGen)

	dto := RegisterUserDTO{
		FirstName: mock.DefaultFirstName,
		LastName:  mock.DefaultLastName,
		Email:     mock.DefaultEmail,
		Password:  mock.DefaultPassword,
	}

	err := uc.Execute(context.Background(), dto)
	if !errors.Is(err, idGenErr) {
		t.Fatalf("expected idGenErr, got: %v", err)
	}
}

func Test_RegisterUserUsecase_PasswordHasherError(t *testing.T) {
	t.Parallel()

	hasherErr := errors.New("hashing failed")
	hasher := mock.NewMockPasswordHasher()
	hasher.HashPasswordFunc = func(password string) (string, error) {
		return "", hasherErr
	}

	uc := NewRegisterUserUsecase(mock.NewMockUsersRepository(), hasher, mock.NewMockIDGenerator())

	dto := RegisterUserDTO{
		FirstName: mock.DefaultFirstName,
		LastName:  mock.DefaultLastName,
		Email:     mock.DefaultEmail,
		Password:  mock.DefaultPassword,
	}

	err := uc.Execute(context.Background(), dto)
	if !errors.Is(err, hasherErr) {
		t.Fatalf("expected hasherErr, got: %v", err)
	}
}

func Test_RegisterUserUsecase_DomainValidationFailure(t *testing.T) {
	t.Parallel()

	uc := NewRegisterUserUsecase(mock.NewMockUsersRepository(), mock.NewMockPasswordHasher(), mock.NewMockIDGenerator())

	tests := []struct {
		name        string
		dto         RegisterUserDTO
		expectedErr error
	}{
		{
			name: "short first name",
			dto: RegisterUserDTO{
				FirstName: "J",
				LastName:  mock.DefaultLastName,
				Email:     mock.DefaultEmail,
				Password:  mock.DefaultPassword,
			},
			expectedErr: domain.ErrUsernameTooShort(domain.FirstName),
		},
		{
			name: "short last name",
			dto: RegisterUserDTO{
				FirstName: mock.DefaultFirstName,
				LastName:  "D",
				Email:     mock.DefaultEmail,
				Password:  mock.DefaultPassword,
			},
			expectedErr: domain.ErrUsernameTooShort(domain.LastName),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := uc.Execute(context.Background(), tt.dto)
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
	repo := mock.NewMockUsersRepository()
	repo.SaveUserFunc = func(ctx context.Context, user *domain.User) error {
		return saveErr
	}

	uc := NewRegisterUserUsecase(repo, mock.NewMockPasswordHasher(), mock.NewMockIDGenerator())

	dto := RegisterUserDTO{
		FirstName: mock.DefaultFirstName,
		LastName:  mock.DefaultLastName,
		Email:     mock.DefaultEmail,
		Password:  mock.DefaultPassword,
	}

	err := uc.Execute(context.Background(), dto)
	if !errors.Is(err, saveErr) {
		t.Fatalf("expected saveErr, got: %v", err)
	}
}

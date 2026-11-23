package identitymocks

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var DefaultID = "usr-1234"
var DefaultFirstName = "John"
var DefaultLastName = "Doe"
var DefaultPhone = "+14155551234"
var DefaultStaffID = "STF-1234"
var DefaultEmail = "john.doe@example.com"
var DefaultPassword = "Password123$"

func ValidUser() *domain.User {
	email, _ := shareddomain.NewEmail(DefaultEmail)
	firstName, _ := domain.NewUsername(DefaultFirstName, domain.FirstName)
	lastName, _ := domain.NewUsername(DefaultLastName, domain.LastName)
	password, _ := domain.NewPassword(DefaultPassword)
	passwordHash, _ := domain.NewPasswordHash(password.Value())

	user, _ := domain.NewUser(DefaultID, firstName.Value(), lastName.Value(), DefaultPhone, DefaultStaffID, email, passwordHash)

	return user
}

type MockUsersRepository struct {
	SaveUserFunc                        func(ctx context.Context, user *domain.User) error
	GetUserByEmailFunc                  func(ctx context.Context, email *shareddomain.Email) (*domain.User, error)
	GetUserByIDFunc                     func(ctx context.Context, id string) (*domain.User, error)
	CheckUserStatusFunc                 func(ctx context.Context, id string) (domain.Status, error)
	UpdateUserStatusFunc                func(ctx context.Context, user *domain.User) error
	UpdateUserPasswordFunc              func(ctx context.Context, user *domain.User) error
	UpdateUserFunc                      func(ctx context.Context, user *domain.User) error
	ListUsersFunc                       func(ctx context.Context, filter domain.UserListFilter) ([]domain.UserSummary, error)
	CountUsersFunc                      func(ctx context.Context, filter domain.UserListFilter) (int, error)
	GetUserDetailFunc                   func(ctx context.Context, id string) (*domain.UserDetail, error)
	GetUserByIDForUpdateFunc            func(ctx context.Context, id string) (*domain.User, error)
	CountOtherActiveAdminsForUpdateFunc func(ctx context.Context, id string) (int, error)
	UpdateUserRoleFunc                  func(ctx context.Context, user *domain.User) error
}

func NewMockUsersRepository() *MockUsersRepository {
	return &MockUsersRepository{
		SaveUserFunc: func(ctx context.Context, user *domain.User) error {
			return nil
		},
		GetUserByEmailFunc: func(ctx context.Context, email *shareddomain.Email) (*domain.User, error) {
			return nil, nil
		},
		GetUserByIDFunc: func(ctx context.Context, id string) (*domain.User, error) {
			return nil, nil
		},
		CheckUserStatusFunc: func(ctx context.Context, id string) (domain.Status, error) {
			return domain.ACTIVE, nil
		},
		UpdateUserStatusFunc: func(ctx context.Context, user *domain.User) error {
			return nil
		},
		UpdateUserPasswordFunc: func(ctx context.Context, user *domain.User) error {
			return nil
		},
		UpdateUserFunc: func(ctx context.Context, user *domain.User) error {
			return nil
		},
		ListUsersFunc: func(ctx context.Context, filter domain.UserListFilter) ([]domain.UserSummary, error) {
			return []domain.UserSummary{}, nil
		},
		CountUsersFunc: func(ctx context.Context, filter domain.UserListFilter) (int, error) {
			return 0, nil
		},
		GetUserDetailFunc: func(ctx context.Context, id string) (*domain.UserDetail, error) {
			return nil, customerrors.NewNotFoundError("user not found")
		},
		GetUserByIDForUpdateFunc: func(ctx context.Context, id string) (*domain.User, error) {
			return ValidUser(), nil
		},
		CountOtherActiveAdminsForUpdateFunc: func(ctx context.Context, id string) (int, error) {
			return 1, nil
		},
		UpdateUserRoleFunc: func(ctx context.Context, user *domain.User) error {
			return nil
		},
	}
}

func (m *MockUsersRepository) SaveUser(ctx context.Context, user *domain.User) error {
	return m.SaveUserFunc(ctx, user)
}

func (m *MockUsersRepository) GetUserByEmail(ctx context.Context, email *shareddomain.Email) (*domain.User, error) {
	return m.GetUserByEmailFunc(ctx, email)
}

func (m *MockUsersRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return m.GetUserByIDFunc(ctx, id)
}

func (m *MockUsersRepository) CheckUserStatus(ctx context.Context, id string) (domain.Status, error) {
	return m.CheckUserStatusFunc(ctx, id)
}

func (m *MockUsersRepository) UpdateUserStatus(ctx context.Context, user *domain.User) error {
	return m.UpdateUserStatusFunc(ctx, user)
}

func (m *MockUsersRepository) UpdateUserPassword(ctx context.Context, user *domain.User) error {
	return m.UpdateUserPasswordFunc(ctx, user)
}

func (m *MockUsersRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	return m.UpdateUserFunc(ctx, user)
}

func (m *MockUsersRepository) ListUsers(ctx context.Context, filter domain.UserListFilter) ([]domain.UserSummary, error) {
	return m.ListUsersFunc(ctx, filter)
}

func (m *MockUsersRepository) CountUsers(ctx context.Context, filter domain.UserListFilter) (int, error) {
	return m.CountUsersFunc(ctx, filter)
}

func (m *MockUsersRepository) GetUserDetail(ctx context.Context, id string) (*domain.UserDetail, error) {
	return m.GetUserDetailFunc(ctx, id)
}

func (m *MockUsersRepository) GetUserByIDForUpdate(ctx context.Context, id string) (*domain.User, error) {
	return m.GetUserByIDForUpdateFunc(ctx, id)
}

func (m *MockUsersRepository) CountOtherActiveAdminsForUpdate(ctx context.Context, id string) (int, error) {
	return m.CountOtherActiveAdminsForUpdateFunc(ctx, id)
}

func (m *MockUsersRepository) UpdateUserRole(ctx context.Context, user *domain.User) error {
	return m.UpdateUserRoleFunc(ctx, user)
}

// Compile-time check that MockUsersRepository implements UsersRepository.
var _ identityout.UsersRepository = (*MockUsersRepository)(nil)

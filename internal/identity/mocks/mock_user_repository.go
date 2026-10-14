package identitymocks

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

var DefaultID = "usr-1234"
var DefaultFirstName = "John"
var DefaultLastName = "Doe"
var DefaultEmail = "john.doe@example.com"
var DefaultPassword = "Password123$"

func ValidUser() *domain.User {
	email, _ := shareddomain.NewEmail(DefaultEmail)
	firstName, _ := domain.NewUsername(DefaultFirstName, domain.FirstName)
	lastName, _ := domain.NewUsername(DefaultLastName, domain.LastName)
	password, _ := domain.NewPassword(DefaultPassword)
	passwordHash, _ := domain.NewPasswordHash(password.Value())

	user, _ := domain.NewUser(DefaultID, firstName.Value(), lastName.Value(), email, passwordHash)

	return user
}

type MockUsersRepository struct {
	SaveUserFunc           func(ctx context.Context, user *domain.User) error
	GetUserByEmailFunc     func(ctx context.Context, email *shareddomain.Email) (*domain.User, error)
	GetUserByIDFunc        func(ctx context.Context, id string) (*domain.User, error)
	UpdateUserStatusFunc   func(ctx context.Context, user *domain.User) error
	UpdateUserPasswordFunc func(ctx context.Context, user *domain.User) error
	UpdateUserFunc         func(ctx context.Context, user *domain.User) error
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
		UpdateUserStatusFunc: func(ctx context.Context, user *domain.User) error {
			return nil
		},
		UpdateUserPasswordFunc: func(ctx context.Context, user *domain.User) error {
			return nil
		},
		UpdateUserFunc: func(ctx context.Context, user *domain.User) error {
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

func (m *MockUsersRepository) UpdateUserStatus(ctx context.Context, user *domain.User) error {
	return m.UpdateUserStatusFunc(ctx, user)
}

func (m *MockUsersRepository) UpdateUserPassword(ctx context.Context, user *domain.User) error {
	return m.UpdateUserPasswordFunc(ctx, user)
}

func (m *MockUsersRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	return m.UpdateUserFunc(ctx, user)
}

// Compile-time check that MockUsersRepository implements UsersRepository.
var _ identityout.UsersRepository = (*MockUsersRepository)(nil)

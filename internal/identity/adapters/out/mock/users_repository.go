package mock

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

var _ identityout.UsersRepository = (*MockUsersRepository)(nil)

type MockUsersRepository struct {
	SaveUserFunc           func(ctx context.Context, user *domain.User) error
	GetUserByEmailFunc     func(ctx context.Context, email *domain.Email) (*domain.User, error)
	GetUserByIDFunc        func(ctx context.Context, id string) (*domain.User, error)
	UpdateUserStatusFunc   func(ctx context.Context, user *domain.User) error
	UpdateUserPasswordFunc func(ctx context.Context, user *domain.User) error
	UpdateUserFunc         func(ctx context.Context, user *domain.User) error
}

func NewMockUsersRepository() *MockUsersRepository {
	return &MockUsersRepository{}
}

func (m *MockUsersRepository) SaveUser(ctx context.Context, user *domain.User) error {
	if m.SaveUserFunc != nil {
		return m.SaveUserFunc(ctx, user)
	}
	return nil
}

func (m *MockUsersRepository) GetUserByEmail(ctx context.Context, email *domain.Email) (*domain.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockUsersRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUsersRepository) UpdateUserStatus(ctx context.Context, user *domain.User) error {
	if m.UpdateUserStatusFunc != nil {
		return m.UpdateUserStatusFunc(ctx, user)
	}
	return nil
}

func (m *MockUsersRepository) UpdateUserPassword(ctx context.Context, user *domain.User) error {
	if m.UpdateUserPasswordFunc != nil {
		return m.UpdateUserPasswordFunc(ctx, user)
	}
	return nil
}

func (m *MockUsersRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, user)
	}
	return nil
}


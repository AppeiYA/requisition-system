package mock

import (
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

var _ identityout.PasswordHasher = (*MockPasswordHasher)(nil)

type MockPasswordHasher struct {
	HashPasswordFunc    func(password string) (string, error)
	ComparePasswordFunc func(hashedPassword, password string) error
}

func NewMockPasswordHasher() *MockPasswordHasher {
	return &MockPasswordHasher{}
}

func (m *MockPasswordHasher) HashPassword(password string) (string, error) {
	if m.HashPasswordFunc != nil {
		return m.HashPasswordFunc(password)
	}
	return "hashed_" + password, nil
}

func (m *MockPasswordHasher) ComparePassword(hashedPassword, password string) error {
	if m.ComparePasswordFunc != nil {
		return m.ComparePasswordFunc(hashedPassword, password)
	}
	return nil
}


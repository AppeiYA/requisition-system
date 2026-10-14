package identitymocks

import (
	"fmt"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

type MockPasswordHasher struct {
	HashPasswordFunc    func(password string) (string, error)
	ComparePasswordFunc func(hashedPassword, password string) error
}

func NewMockPasswordHasher() *MockPasswordHasher {
	return &MockPasswordHasher{
		HashPasswordFunc: func(password string) (string, error) {
			return password, nil
		},
		ComparePasswordFunc: func(hashedPassword, password string) error {
			if hashedPassword != password {
				return fmt.Errorf("invalid password")
			}
			return nil
		},
	}
}

func (m *MockPasswordHasher) HashPassword(password string) (string, error) {
	return m.HashPasswordFunc(password)
}

func (m *MockPasswordHasher) ComparePassword(hashedPassword, password string) error {
	return m.ComparePasswordFunc(hashedPassword, password)
}

var _ identityout.PasswordHasher = (*MockPasswordHasher)(nil)

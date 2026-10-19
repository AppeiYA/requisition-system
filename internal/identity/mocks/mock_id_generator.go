package identitymocks

import identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"

type MockIDGenerator struct {
	GenerateIDFunc func(prefix string) (string, error)
}

func NewMockIDGenerator() *MockIDGenerator {
	return &MockIDGenerator{
		GenerateIDFunc: func(prefix string) (string, error) {
			return prefix + "123", nil
		},
	}
}

func (m *MockIDGenerator) GenerateID(prefix string) (string, error) {
	return m.GenerateIDFunc(prefix)
}

var _ identityout.IDGenerator = (*MockIDGenerator)(nil)

package mock

import (
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

var _ identityout.IDGenerator = (*MockIDGenerator)(nil)

type MockIDGenerator struct {
	GenerateIDFunc func(prefix string) (string, error)
}

func NewMockIDGenerator() *MockIDGenerator {
	return &MockIDGenerator{}
}

func (m *MockIDGenerator) GenerateID(prefix string) (string, error) {
	if m.GenerateIDFunc != nil {
		return m.GenerateIDFunc(prefix)
	}
	return prefix + "-12345678", nil
}


package invitationmocks

import invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"

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

var _ invitationout.IDGenerator = (*MockIDGenerator)(nil)

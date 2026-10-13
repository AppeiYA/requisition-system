package identitymocks

import identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"

type MockRefreshTokenGenerator struct {
	GenerateFunc func() (string, error)
	HashFunc     func(token string) string
}

func NewMockRefreshTokenGenerator() *MockRefreshTokenGenerator {
	return &MockRefreshTokenGenerator{
		GenerateFunc: func() (string, error) {
			return "refresh-token", nil
		},
		HashFunc: func(token string) string {
			return "hash:" + token
		},
	}
}

func (m *MockRefreshTokenGenerator) Generate() (string, error) {
	return m.GenerateFunc()
}

func (m *MockRefreshTokenGenerator) Hash(token string) string {
	return m.HashFunc(token)
}

var _ identityout.RefreshTokenGenerator = (*MockRefreshTokenGenerator)(nil)

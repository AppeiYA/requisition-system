package identitymocks

import (
	"context"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

type MockTokenService struct {
	GenerateTokenFunc func(ctx context.Context, user *identityout.TokenModel) (string, error)
	ValidateTokenFunc func(ctx context.Context, token string) (*identityout.TokenModel, error)
}

func NewMockTokenService() *MockTokenService {
	return &MockTokenService{}
}

func (m *MockTokenService) GenerateToken(ctx context.Context, user *identityout.TokenModel) (string, error) {
	if m.GenerateTokenFunc != nil {
		return m.GenerateTokenFunc(ctx, user)
	}
	return "", nil
}

func (m *MockTokenService) ValidateToken(ctx context.Context, token string) (*identityout.TokenModel, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(ctx, token)
	}
	return nil, nil
}

var _ identityout.TokenService = (*MockTokenService)(nil)

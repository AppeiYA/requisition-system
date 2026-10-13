package identitymocks

import (
	"context"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

type MockAccessTokenService struct {
	GenerateAccessTokenFunc func(ctx context.Context, user *identityout.TokenModel) (*identityout.AccessToken, error)
	ValidateAccessTokenFunc func(ctx context.Context, token string) (*identityout.TokenModel, error)
}

func NewMockAccessTokenService() *MockAccessTokenService {
	return &MockAccessTokenService{
		GenerateAccessTokenFunc: func(ctx context.Context, user *identityout.TokenModel) (*identityout.AccessToken, error) {
			return &identityout.AccessToken{Value: "access-token", ExpiresIn: 900}, nil
		},
		ValidateAccessTokenFunc: func(ctx context.Context, token string) (*identityout.TokenModel, error) {
			return &identityout.TokenModel{ID: DefaultID}, nil
		},
	}
}

func (m *MockAccessTokenService) GenerateAccessToken(ctx context.Context, user *identityout.TokenModel) (*identityout.AccessToken, error) {
	return m.GenerateAccessTokenFunc(ctx, user)
}

func (m *MockAccessTokenService) ValidateAccessToken(ctx context.Context, token string) (*identityout.TokenModel, error) {
	return m.ValidateAccessTokenFunc(ctx, token)
}

var _ identityout.AccessTokenService = (*MockAccessTokenService)(nil)

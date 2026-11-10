package identitymocks

import (
	"context"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

type MockRefreshSessionRepository struct {
	CreateFunc func(ctx context.Context, userID, tokenHash string) error
	RotateFunc func(ctx context.Context, oldTokenHash, newTokenHash string) (*identityout.RefreshSession, error)
	RevokeFunc func(ctx context.Context, tokenHash string) error
}

func NewMockRefreshSessionRepository() *MockRefreshSessionRepository {
	return &MockRefreshSessionRepository{
		CreateFunc: func(ctx context.Context, userID, tokenHash string) error {
			return nil
		},
		RotateFunc: func(ctx context.Context, oldTokenHash, newTokenHash string) (*identityout.RefreshSession, error) {
			return &identityout.RefreshSession{
				UserID:   DefaultID,
				FamilyID: "family-id",
			}, nil
		},
		RevokeFunc: func(ctx context.Context, tokenHash string) error {
			return nil
		},
	}
}

func (m *MockRefreshSessionRepository) Create(ctx context.Context, userID, tokenHash string) error {
	return m.CreateFunc(ctx, userID, tokenHash)
}

func (m *MockRefreshSessionRepository) Rotate(ctx context.Context, oldTokenHash, newTokenHash string) (*identityout.RefreshSession, error) {
	return m.RotateFunc(ctx, oldTokenHash, newTokenHash)
}

func (m *MockRefreshSessionRepository) Revoke(ctx context.Context, tokenHash string) error {
	return m.RevokeFunc(ctx, tokenHash)
}

var _ identityout.RefreshSessionRepository = (*MockRefreshSessionRepository)(nil)

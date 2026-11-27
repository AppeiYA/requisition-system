package identitymocks

import (
	"context"
	"time"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

type MockPasswordResetTokenRepository struct {
	CreateFunc                  func(ctx context.Context, token *domain.PasswordResetToken) error
	GetByTokenHashFunc          func(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error)
	GetLatestActiveForUserFunc  func(ctx context.Context, userID string, now time.Time) (*domain.PasswordResetToken, error)
	CountCreatedSinceFunc       func(ctx context.Context, userID string, since time.Time) (int, error)
	InvalidateActiveForUserFunc func(ctx context.Context, userID string, at time.Time) error
	InvalidateByIDFunc          func(ctx context.Context, id string, at time.Time) error
	ConsumeFunc                 func(ctx context.Context, tokenHash, userID string, now time.Time) (*domain.PasswordResetToken, error)
	DeleteOlderThanFunc         func(ctx context.Context, before time.Time) (int64, error)
}

func NewMockPasswordResetTokenRepository() *MockPasswordResetTokenRepository {
	return &MockPasswordResetTokenRepository{
		CreateFunc: func(ctx context.Context, token *domain.PasswordResetToken) error {
			return nil
		},
		GetByTokenHashFunc: func(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
			return nil, identityout.ErrPasswordResetTokenNotFound
		},
		GetLatestActiveForUserFunc: func(ctx context.Context, userID string, now time.Time) (*domain.PasswordResetToken, error) {
			return nil, nil
		},
		CountCreatedSinceFunc: func(ctx context.Context, userID string, since time.Time) (int, error) {
			return 0, nil
		},
		InvalidateActiveForUserFunc: func(ctx context.Context, userID string, at time.Time) error {
			return nil
		},
		InvalidateByIDFunc: func(ctx context.Context, id string, at time.Time) error {
			return nil
		},
		ConsumeFunc: func(ctx context.Context, tokenHash, userID string, now time.Time) (*domain.PasswordResetToken, error) {
			return nil, identityout.ErrPasswordResetTokenNotFound
		},
		DeleteOlderThanFunc: func(ctx context.Context, before time.Time) (int64, error) {
			return 0, nil
		},
	}
}

func (m *MockPasswordResetTokenRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	return m.CreateFunc(ctx, token)
}

func (m *MockPasswordResetTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	return m.GetByTokenHashFunc(ctx, tokenHash)
}

func (m *MockPasswordResetTokenRepository) GetLatestActiveForUser(ctx context.Context, userID string, now time.Time) (*domain.PasswordResetToken, error) {
	return m.GetLatestActiveForUserFunc(ctx, userID, now)
}

func (m *MockPasswordResetTokenRepository) CountCreatedSince(ctx context.Context, userID string, since time.Time) (int, error) {
	return m.CountCreatedSinceFunc(ctx, userID, since)
}

func (m *MockPasswordResetTokenRepository) InvalidateActiveForUser(ctx context.Context, userID string, at time.Time) error {
	return m.InvalidateActiveForUserFunc(ctx, userID, at)
}

func (m *MockPasswordResetTokenRepository) InvalidateByID(ctx context.Context, id string, at time.Time) error {
	return m.InvalidateByIDFunc(ctx, id, at)
}

func (m *MockPasswordResetTokenRepository) Consume(ctx context.Context, tokenHash, userID string, now time.Time) (*domain.PasswordResetToken, error) {
	return m.ConsumeFunc(ctx, tokenHash, userID, now)
}

func (m *MockPasswordResetTokenRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	return m.DeleteOlderThanFunc(ctx, before)
}

var _ identityout.PasswordResetTokenRepository = (*MockPasswordResetTokenRepository)(nil)

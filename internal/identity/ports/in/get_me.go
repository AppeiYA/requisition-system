package identityin

import (
	"context"
	"time"
)

type UserResponse struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type GetMeUsecaseInt interface {
	Execute(ctx context.Context, userID string) (*UserResponse, error)
}

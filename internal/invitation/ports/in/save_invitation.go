package invitationin

import (
	"context"
	"time"
)

type SaveInvitationResponse struct {
	ID        string
	Token     string
	ExpiresAt time.Time
}

type SaveInvitationUsecaseInt interface {
	Execute(ctx context.Context, email string, inviterID string) (*SaveInvitationResponse, error)
}

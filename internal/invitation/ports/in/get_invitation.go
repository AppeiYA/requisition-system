package invitationin

import (
	"context"
	"time"
)

type GetInvitationOutput struct {
	ID           string
	FirstName    string
	LastName     string
	Phone        string
	StaffID      string
	DepartmentID string
	JobTitle     string
	Email       	string
	InvitedBy    string
	ExpiresAt    time.Time
	AcceptedAt   *time.Time
	RevokedAt    *time.Time
	CreatedAt    time.Time
}

type GetInvitationUsecaseInt interface {
	ByEmail(ctx context.Context, email string) (*GetInvitationOutput, error)
}
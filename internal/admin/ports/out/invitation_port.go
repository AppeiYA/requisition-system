package adminout

import (
	"context"
	"time"

	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

type SavedInvitation struct {
	ID        string
	Token     string
	ExpiresAt time.Time
}

type SaveInvitationParams struct {
	Email        *shareddomain.Email
	FirstName    string
	LastName     string
	Phone        string
	StaffID      string
	DepartmentID string
	JobTitle     string
	InviterID    string
}

type AdminInvitationPort interface {
	SaveInvitation(ctx context.Context, params SaveInvitationParams) (*SavedInvitation, error)
}

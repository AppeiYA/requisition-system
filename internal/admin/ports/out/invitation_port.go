package adminout

import (
	"context"
	"time"

	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

// SavedInvitation is the result of persisting an invitation. The token is the
// raw, single-use token that must be delivered to the invitee (only its hash is
// stored).
type SavedInvitation struct {
	ID        string
	Token     string
	ExpiresAt time.Time
}

// SaveInvitationParams carries the invitation details, including the
// admin-provided prefill values for the invitee and the department assignment
// the invitee will receive on acceptance.
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
	CheckActiveInvitationExists(ctx context.Context, email *shareddomain.Email) (bool, error)
	SaveInvitation(ctx context.Context, params SaveInvitationParams) (*SavedInvitation, error)
}

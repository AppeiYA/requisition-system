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

// InvitationListFilter is the admin module's transport-free view of the
// invitation filters. The external adapter translates it into the invitation
// module's own filter type.
type InvitationListFilter struct {
	Search       string
	Status       string
	DepartmentID string
	InvitedBy    string
	From         *time.Time
	To           *time.Time
	Limit        int
	Offset       int
}

type InvitationSummary struct {
	ID               string
	FirstName        string
	LastName         string
	Email            string
	DepartmentID     string
	DepartmentName   string
	JobTitle         string
	Status           string
	InvitedBy        string
	InviterName      string
	ExpiresAt        time.Time
	AcceptedAt       *time.Time
	RevokedAt        *time.Time
	DeliveredAt      *time.Time
	DeliveryFailedAt *time.Time
	CreatedAt        time.Time
}

type ResendInvitationParams struct {
	InvitationID string
	ActorID      string
}

// ResendInvitationResult carries the rotated raw token to the admin module so
// it can queue the replacement email. Token hashes never leave the invitation
// module and are never written to the audit log.
type ResendInvitationResult struct {
	ID           string
	Token        string
	Email        string
	FirstName    string
	LastName     string
	InvitedBy    string
	DepartmentID string
	JobTitle     string
	ExpiresAt    time.Time
}

type RevokeInvitationParams struct {
	InvitationID string
	ActorID      string
	Reason       string
}

type AdminInvitationPort interface {
	SaveInvitation(ctx context.Context, params SaveInvitationParams) (*SavedInvitation, error)
}

// AdminInvitationManagementPort is the admin module's read-and-mutate view of
// invitations. It is separate from the creation port so the invite use case can
// keep its narrow dependency.
type AdminInvitationManagementPort interface {
	ListInvitations(ctx context.Context, filter InvitationListFilter) ([]InvitationSummary, int, error)
	GetInvitation(ctx context.Context, invitationID string) (*InvitationSummary, error)
	ResendInvitation(ctx context.Context, params ResendInvitationParams) (*ResendInvitationResult, error)
	RevokeInvitation(ctx context.Context, params RevokeInvitationParams) error
}

package invitationin

import (
	"context"
	"time"

	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
)

// AdminInvitationListFilter is the invitation module's own view of the admin
// list filters. It carries no transport tags; the HTTP adapter translates query
// parameters into it.
// AdminInvitationsPortIn groups the invitation use cases used by the admin
// module. It is kept separate from InvitationPortIn so existing constructors
// do not have to change for every new admin capability.
type AdminInvitationsPortIn struct {
	ListInvitations  AdminListInvitationsUsecaseInt
	GetInvitation    AdminGetInvitationUsecaseInt
	ResendInvitation AdminResendInvitationUsecaseInt
	RevokeInvitation AdminRevokeInvitationUsecaseInt
}

type AdminInvitationListFilter struct {
	Search       string
	Status       string
	DepartmentID string
	InvitedBy    string
	From         *time.Time
	To           *time.Time
	Limit        int
	Offset       int
}

type AdminListInvitationsResult struct {
	Invitations []invitationdomain.InvitationSummary
	Total       int
}

type AdminListInvitationsUsecaseInt interface {
	Execute(ctx context.Context, filter AdminInvitationListFilter) (*AdminListInvitationsResult, error)
}

type AdminGetInvitationUsecaseInt interface {
	Execute(ctx context.Context, invitationID string) (*invitationdomain.InvitationSummary, error)
}

type ResendInvitationCommand struct {
	InvitationID string
	ActorID      string
}

// ResendInvitationResult includes the raw token so the admin module can queue
// the replacement email. It is never persisted in this form and never written
// to the audit log.
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

type AdminResendInvitationUsecaseInt interface {
	Execute(ctx context.Context, cmd ResendInvitationCommand) (*ResendInvitationResult, error)
}

type RevokeInvitationCommand struct {
	InvitationID string
	ActorID      string
	Reason       string
}

type AdminRevokeInvitationUsecaseInt interface {
	Execute(ctx context.Context, cmd RevokeInvitationCommand) error
}

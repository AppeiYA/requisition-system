package adminin

import (
	"context"
	"time"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
)

// ValidInvitationStatuses is the invitation status allow-list for the HTTP
// filter. The admin layer owns its own vocabulary; the invitation module
// validates the value again when it receives the filter.
var ValidInvitationStatuses = map[string]bool{
	"pending":   true,
	"delivered": true,
	"failed":    true,
	"accepted":  true,
	"revoked":   true,
	"expired":   true,
}

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

type ListInvitationsResult struct {
	Invitations []adminout.InvitationSummary
	Total       int
}

type ListInvitationsUseCase interface {
	Execute(ctx context.Context, filter InvitationListFilter) (*ListInvitationsResult, error)
}

type GetInvitationUseCase interface {
	Execute(ctx context.Context, invitationID string) (*adminout.InvitationSummary, error)
}

type ResendInvitationCommand struct {
	InvitationID string
	ActorID      string
}

type ResendInvitationUseCase interface {
	Execute(ctx context.Context, cmd ResendInvitationCommand) (*adminout.ResendInvitationResult, error)
}

type RevokeInvitationCommand struct {
	InvitationID string
	ActorID      string
	Reason       string
}

type RevokeInvitationUseCase interface {
	Execute(ctx context.Context, cmd RevokeInvitationCommand) error
}

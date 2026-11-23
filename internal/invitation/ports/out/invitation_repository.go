package invitationout

import (
	"context"
	"time"

	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
)

type InvitationRepository interface {
	SaveInvitation(ctx context.Context, invitation *invitationdomain.Invitation) error
	GetInvitationByTokenHash(ctx context.Context, tokenHash string) (*invitationdomain.Invitation, error)
	LockInvitationByTokenHash(ctx context.Context, tokenHash string) (*invitationdomain.Invitation, error)
	GetActiveInvitationByEmail(ctx context.Context, email string) (*invitationdomain.Invitation, error)
	MarkInvitationAccepted(ctx context.Context, invitationID string) error
	RevokeInvitation(ctx context.Context, invitationID string) error
	MarkInvitationDelivered(ctx context.Context, invitationID string) error
	MarkInvitationDeliveryFailed(ctx context.Context, invitationID string) error
	GetInvitationByID(ctx context.Context, invitationID string) (*invitationdomain.Invitation, error)
	GetInvitationByEmail(ctx context.Context, email string) (*invitationdomain.Invitation, error)
	UpdateInvitationStatus(ctx context.Context, invitationID string, status invitationdomain.InvitationStatus) error
	ExpireDueInvitations(ctx context.Context, now time.Time) error

	// Admin additions. LockInvitationByID is used by the resend and revoke use
	// cases so state checks and the following write are one atomic operation.
	LockInvitationByID(ctx context.Context, invitationID string) (*invitationdomain.Invitation, error)
	ListInvitations(ctx context.Context, filter invitationdomain.InvitationListFilter) ([]invitationdomain.InvitationSummary, error)
	CountInvitations(ctx context.Context, filter invitationdomain.InvitationListFilter) (int, error)
	GetInvitationSummary(ctx context.Context, invitationID string) (*invitationdomain.InvitationSummary, error)
	RotateInvitationToken(ctx context.Context, invitationID, tokenHash string, expiresAt time.Time) error
}

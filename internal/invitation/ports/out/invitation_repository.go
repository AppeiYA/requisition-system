package invitationout

import (
	"context"
	"time"

	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
)

type InvitationRepository interface {
	SaveInvitation(ctx context.Context, invitation *invitationdomain.Invitation) error
	GetInvitationByTokenHash(ctx context.Context, tokenHash string) (*invitationdomain.Invitation, error)
	GetActiveInvitationByEmail(ctx context.Context, email string) (*invitationdomain.Invitation, error)
	MarkInvitationAccepted(ctx context.Context, invitationID string) error
	RevokeInvitation(ctx context.Context, invitationID string) error
	GetInvitationByID(ctx context.Context, invitationID string) (*invitationdomain.Invitation, error)
	GetInvitationByEmail(ctx context.Context, email string) (*invitationdomain.Invitation, error)
	UpdateInvitationStatus(ctx context.Context, invitationID string, status invitationdomain.InvitationStatus) error
	ExpireDueInvitations(ctx context.Context, now time.Time) error
}

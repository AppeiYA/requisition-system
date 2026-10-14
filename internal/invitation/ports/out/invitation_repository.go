package invitationout

import (
	"context"

	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
)

type InvitationRepository interface {
	SaveInvitation(ctx context.Context, invitation *invitationdomain.Invitation) error
	GetInvitationByTokenHash(ctx context.Context, tokenHash string) (*invitationdomain.Invitation, error)
	GetActiveInvitationByEmail(ctx context.Context, email string) (*invitationdomain.Invitation, error)
	MarkInvitationAccepted(ctx context.Context, invitationID string) error
	RevokeInvitation(ctx context.Context, invitationID string) error
}

package usecase

import (
	"time"

	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

const (
	testInvitationEmail = "invitee@example.com"
	testInvitationToken = "raw-invitation-token"
	testInvitationID    = "inv-1234"
	testInviterID       = "usr-1234"
)

func pendingInvitation() *invitationdomain.Invitation {
	email, _ := shareddomain.NewEmail(testInvitationEmail)
	return invitationdomain.NewInvitation(
		testInvitationID,
		email,
		invitationdomain.HashInvitationToken(testInvitationToken),
		testInviterID,
		time.Now().Add(time.Hour),
		time.Now(),
	)
}

func invitationWithState(expiresAt time.Time, acceptedAt, revokedAt *time.Time) *invitationdomain.Invitation {
	inv, _ := invitationdomain.ReconstituteInvitation(
		testInvitationID,
		testInvitationEmail,
		invitationdomain.HashInvitationToken(testInvitationToken),
		testInviterID,
		expiresAt,
		acceptedAt,
		revokedAt,
		time.Now(),
	)
	return inv
}

package usecase

import (
	"time"

	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

const (
	testInvitationEmail     = "invitee@example.com"
	testInvitationToken     = "raw-invitation-token"
	testInvitationID        = "inv-1234"
	testInviterID           = "usr-1234"
	testInvitationFirstName = "Jane"
	testInvitationLastName  = "Doe"
	testInvitationPhone     = "+14155551234"
	testInvitationStaffID   = "STF-1234"
	testDepartmentID        = "dpt-1234"
	testJobTitle            = "Software Engineer"
	testRegisteredUserID    = "usr-5678"
)

func pendingInvitation() *invitationdomain.Invitation {
	email, _ := shareddomain.NewEmail(testInvitationEmail)
	firstName, _ := shareddomain.NewUsername(testInvitationFirstName, shareddomain.FirstName)
	lastName, _ := shareddomain.NewUsername(testInvitationLastName, shareddomain.LastName)
	phone, _ := shareddomain.NewPhoneNumber(testInvitationPhone)
	invitation, err := invitationdomain.NewInvitation(
		testInvitationID,
		firstName,
		lastName,
		phone,
		testInvitationStaffID,
		testDepartmentID,
		testJobTitle,
		email,
		invitationdomain.HashInvitationToken(testInvitationToken),
		testInviterID,
	)
	if err != nil {
		panic(err)
	}
	return invitation
}

func invitationWithState(expiresAt time.Time, acceptedAt, revokedAt *time.Time) *invitationdomain.Invitation {
	inv, _ := invitationdomain.ReconstituteInvitation(
		testInvitationID,
		testInvitationFirstName,
		testInvitationLastName,
		testInvitationPhone,
		testInvitationStaffID,
		testDepartmentID,
		testJobTitle,
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

package models

import (
	"time"

	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
)

type InvitationModel struct {
	ID         string     `db:"id"`
	Email      string     `db:"email"`
	TokenHash  string     `db:"token_hash"`
	InvitedBy  string     `db:"invited_by"`
	ExpiresAt  time.Time  `db:"expires_at"`
	AcceptedAt *time.Time `db:"accepted_at"`
	RevokedAt  *time.Time `db:"revoked_at"`
	CreatedAt  time.Time  `db:"created_at"`
}

func FromDomain(inv *invitationdomain.Invitation) *InvitationModel {
	return &InvitationModel{
		ID:         inv.ID(),
		Email:      inv.Email().Value(),
		TokenHash:  inv.TokenHash(),
		InvitedBy:  inv.InvitedBy(),
		ExpiresAt:  inv.ExpiresAt(),
		AcceptedAt: inv.AcceptedAt(),
		RevokedAt:  inv.RevokedAt(),
		CreatedAt:  inv.CreatedAt(),
	}
}

func (m InvitationModel) ToDomain() (*invitationdomain.Invitation, error) {
	return invitationdomain.ReconstituteInvitation(
		m.ID,
		m.Email,
		m.TokenHash,
		m.InvitedBy,
		m.ExpiresAt,
		m.AcceptedAt,
		m.RevokedAt,
		m.CreatedAt,
	)
}

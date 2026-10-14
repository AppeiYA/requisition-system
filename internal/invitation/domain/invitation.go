package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrInvitationInvalidToken  = customerrors.NewBadRequestError("invalid invitation token")
	ErrInvitationExpired       = customerrors.NewBadRequestError("invitation has expired")
	ErrInvitationRevoked       = customerrors.NewBadRequestError("invitation has been revoked")
	ErrInvitationAlreadyUsed   = customerrors.NewBadRequestError("invitation has already been accepted")
	ErrInvitationNotFound      = customerrors.NewNotFoundError("invitation not found")
	ErrInvitationTokenRequired = customerrors.NewBadRequestError("invitation token is required")
)

var InvitationIDPrefix = "inv"

const InvitationTokenBytes = 32
const InvitationExpirationDuration = 24 * time.Hour

type InvitationToken string

func HashInvitationToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

type Invitation struct {
	id         string
	email      *shareddomain.Email
	tokenHash  string
	invitedBy  string
	expiresAt  time.Time
	acceptedAt *time.Time
	revokedAt  *time.Time
	createdAt  time.Time
}

func NewInvitation(id string, email *shareddomain.Email, tokenHash string, invitedBy string, expiresAt time.Time, createdAt time.Time) *Invitation {
	return &Invitation{
		id:        id,
		email:     email,
		tokenHash: tokenHash,
		invitedBy: invitedBy,
		expiresAt: expiresAt,
		createdAt: createdAt,
	}
}

func ReconstituteInvitation(id string, email string, tokenHash string, invitedBy string, expiresAt time.Time, acceptedAt *time.Time, revokedAt *time.Time, createdAt time.Time) (*Invitation, error) {
	validEmail, err := shareddomain.NewEmail(email)
	if err != nil {
		return nil, err
	}
	return &Invitation{
		id:         id,
		email:      validEmail,
		tokenHash:  tokenHash,
		invitedBy:  invitedBy,
		expiresAt:  expiresAt,
		acceptedAt: acceptedAt,
		revokedAt:  revokedAt,
		createdAt:  createdAt,
	}, nil
}

func (i *Invitation) ID() string {
	return i.id
}

func (i *Invitation) Email() *shareddomain.Email {
	return i.email
}

func (i *Invitation) TokenHash() string {
	return i.tokenHash
}

func (i *Invitation) InvitedBy() string {
	return i.invitedBy
}

func (i *Invitation) ExpiresAt() time.Time {
	return i.expiresAt
}

func (i *Invitation) AcceptedAt() *time.Time {
	return i.acceptedAt
}

func (i *Invitation) RevokedAt() *time.Time {
	return i.revokedAt
}

func (i *Invitation) CreatedAt() time.Time {
	return i.createdAt
}

func (i *Invitation) IsAccepted() bool {
	return i.acceptedAt != nil
}

func (i *Invitation) IsRevoked() bool {
	return i.revokedAt != nil
}

func (i *Invitation) IsExpired(now time.Time) bool {
	return now.After(i.expiresAt) || now.Equal(i.expiresAt)
}

func (i *Invitation) IsActive(now time.Time) bool {
	if i.IsAccepted() || i.IsRevoked() {
		return false
	}
	return now.Before(i.expiresAt)
}

func (i *Invitation) MarkAccepted(now time.Time) {
	if i.acceptedAt == nil {
		i.acceptedAt = &now
	}
}

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
	ErrInvitationIDRequired    = customerrors.NewInternalServerError("invitation ID is required")
	ErrInvitationEmailRequired = customerrors.NewBadRequestError("invitation email is required")
	ErrFirstnameRequired       = customerrors.NewBadRequestError("first name is required")
	ErrLastnameRequired        = customerrors.NewBadRequestError("last name is required")
	ErrInvitedByRequired       = customerrors.NewBadRequestError("invited by is required")
	ErrStaffIDTooLong          = customerrors.NewBadRequestError("staff ID must not exceed 20 characters")
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
	id           string
	firstName    *shareddomain.Username
	lastName     *shareddomain.Username
	phone        *shareddomain.PhoneNumber
	staffID      string
	departmentID string
	jobTitle     string
	email        *shareddomain.Email
	tokenHash    string
	invitedBy    string
	status       InvitationStatus
	expiresAt    time.Time
	acceptedAt   *time.Time
	revokedAt    *time.Time
	createdAt    time.Time
}

func NewInvitation(
	id string,
	firstName *shareddomain.Username,
	lastName *shareddomain.Username,
	phone *shareddomain.PhoneNumber,
	staffID string,
	departmentID string,
	jobTitle string,
	email *shareddomain.Email,
	tokenHash string,
	invitedBy string,
) (*Invitation, error) {
	if id == "" {
		return nil, ErrInvitationIDRequired
	}
	if firstName == nil {
		return nil, ErrFirstnameRequired
	}
	if lastName == nil {
		return nil, ErrLastnameRequired
	}
	if email == nil {
		return nil, ErrInvitationEmailRequired
	}
	if tokenHash == "" {
		return nil, ErrInvitationTokenRequired
	}

	if invitedBy == "" {
		return nil, ErrInvitedByRequired
	}

	if len(staffID) > 20 {
		return nil, ErrStaffIDTooLong
	}

	status := InvitationStatusPending

	now := time.Now()
	expiresAt := now.Add(InvitationExpirationDuration)
	createdAt := now

	return &Invitation{
		id:           id,
		firstName:    firstName,
		lastName:     lastName,
		phone:        phone,
		staffID:      staffID,
		departmentID: departmentID,
		jobTitle:     jobTitle,
		email:        email,
		tokenHash:    tokenHash,
		status:       status,
		invitedBy:    invitedBy,
		expiresAt:    expiresAt,
		createdAt:    createdAt,
	}, nil
}

func ReconstituteInvitation(
	id string,
	firstName string,
	lastName string,
	phone string,
	staffID string,
	departmentID string,
	jobTitle string,
	email string,
	tokenHash string,
	invitedBy string,
	expiresAt time.Time,
	acceptedAt *time.Time,
	revokedAt *time.Time,
	createdAt time.Time,
) (*Invitation, error) {
	status := InvitationStatusPending
	if acceptedAt != nil {
		status = InvitationStatusAccepted
	} else if revokedAt != nil {
		status = InvitationStatusRevoked
	} else if !expiresAt.After(time.Now()) {
		status = InvitationStatusExpired
	}

	return reconstituteInvitationWithStatus(
		id, firstName, lastName, phone, staffID, departmentID, jobTitle,
		email, tokenHash, invitedBy, status, expiresAt, acceptedAt, revokedAt,
		createdAt,
	)
}

func ReconstituteInvitationWithStatus(
	id string,
	firstName string,
	lastName string,
	phone string,
	staffID string,
	departmentID string,
	jobTitle string,
	email string,
	tokenHash string,
	invitedBy string,
	status string,
	expiresAt time.Time,
	acceptedAt *time.Time,
	revokedAt *time.Time,
	createdAt time.Time,
) (*Invitation, error) {
	validStatus, err := NewInvitationStatus(status)
	if err != nil {
		return nil, err
	}
	inv, err := reconstituteInvitationWithStatus(
		id, firstName, lastName, phone, staffID, departmentID, jobTitle,
		email, tokenHash, invitedBy, validStatus, expiresAt, acceptedAt,
		revokedAt, createdAt,
	)
	if err != nil {
		return nil, err
	}
	return inv, nil
}

func reconstituteInvitationWithStatus(
	id string,
	firstName string,
	lastName string,
	phone string,
	staffID string,
	departmentID string,
	jobTitle string,
	email string,
	tokenHash string,
	invitedBy string,
	status InvitationStatus,
	expiresAt time.Time,
	acceptedAt *time.Time,
	revokedAt *time.Time,
	createdAt time.Time,
) (*Invitation, error) {
	validFirstName, err := shareddomain.NewUsername(firstName, shareddomain.FirstName)
	if err != nil {
		return nil, err
	}
	validLastName, err := shareddomain.NewUsername(lastName, shareddomain.LastName)
	if err != nil {
		return nil, err
	}
	validPhone, err := shareddomain.NewOptionalPhoneNumber(phone)
	if err != nil {
		return nil, err
	}
	validEmail, err := shareddomain.NewEmail(email)
	if err != nil {
		return nil, err
	}
	return &Invitation{
		id:           id,
		firstName:    validFirstName,
		lastName:     validLastName,
		phone:        validPhone,
		staffID:      staffID,
		departmentID: departmentID,
		jobTitle:     jobTitle,
		email:        validEmail,
		tokenHash:    tokenHash,
		invitedBy:    invitedBy,
		status:       status,
		expiresAt:    expiresAt,
		acceptedAt:   acceptedAt,
		revokedAt:    revokedAt,
		createdAt:    createdAt,
	}, nil
}

func (i *Invitation) ID() string {
	return i.id
}

func (i *Invitation) FirstName() *shareddomain.Username {
	return i.firstName
}

func (i *Invitation) LastName() *shareddomain.Username {
	return i.lastName
}

func (i *Invitation) Phone() *shareddomain.PhoneNumber {
	return i.phone
}

func (i *Invitation) StaffID() string {
	return i.staffID
}

// DepartmentID is the department the invitee will be assigned to. It is fixed
// when the invitation is created and is immutable afterwards.
func (i *Invitation) DepartmentID() string {
	return i.departmentID
}

// JobTitle is the invitee's role in the assigned department. It is fixed when
// the invitation is created and is immutable afterwards.
func (i *Invitation) JobTitle() string {
	return i.jobTitle
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

func (i *Invitation) Status() InvitationStatus {
	return i.status
}

func (i *Invitation) MarkAccepted() error {
	now := time.Now()
	if !i.status.CanTransitionTo(InvitationStatusAccepted) {
		return ErrInvalidInvitationStatusTransition
	}
	i.status = InvitationStatusAccepted
	i.acceptedAt = &now
	return nil
}

func (i *Invitation) MarkAsDelivered() error {
	if !i.status.CanTransitionTo(InvitationStatusDelivered) {
		return ErrInvalidInvitationStatusTransition
	}
	i.status = InvitationStatusDelivered
	return nil
}

func (i *Invitation) MarkAsFailed() error {
	if !i.status.CanTransitionTo(InvitationStatusFailed) {
		return ErrInvalidInvitationStatusTransition
	}
	i.status = InvitationStatusFailed
	return nil
}

func (i *Invitation) MarkAsExpired() error {
	if !i.status.CanTransitionTo(InvitationStatusExpired) {
		return ErrInvalidInvitationStatusTransition
	}
	i.status = InvitationStatusExpired
	return nil
}

func (i *Invitation) MarkAsRevoked() error {
	if !i.status.CanTransitionTo(InvitationStatusRevoked) {
		return ErrInvalidInvitationStatusTransition
	}
	i.status = InvitationStatusRevoked
	return nil
}

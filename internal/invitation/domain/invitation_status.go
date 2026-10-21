package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

type InvitationStatus string
const (
	InvitationStatusPending InvitationStatus = "pending"
	InvitationStatusDelivered InvitationStatus = "delivered"
	InvitationStatusFailed   InvitationStatus = "failed"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusRevoked  InvitationStatus = "revoked"
	InvitationStatusExpired  InvitationStatus = "expired"
)
var (
	ErrInvalidInvitationStatus = customerrors.NewBadRequestError("invalid invitation status")
	ErrInvalidInvitationStatusTransition = customerrors.NewBadRequestError("invalid invitation status transition")
)

var validInvitationStatusTransitions = map[InvitationStatus]map[InvitationStatus]struct{}{
	InvitationStatusPending: {
		InvitationStatusDelivered: {},
		InvitationStatusAccepted:  {},
		InvitationStatusRevoked:   {},
		InvitationStatusExpired:   {},
		InvitationStatusFailed:    {},
	},
	InvitationStatusDelivered: {
		InvitationStatusAccepted: {},
		InvitationStatusRevoked:  {},
		InvitationStatusExpired:  {},
	},
	InvitationStatusFailed: {},
	InvitationStatusAccepted: {},
	InvitationStatusRevoked:  {},
	InvitationStatusExpired:  {},
}

func (s InvitationStatus) IsValid() bool {
	switch s {
	case InvitationStatusPending, InvitationStatusDelivered, InvitationStatusFailed, InvitationStatusAccepted, InvitationStatusRevoked, InvitationStatusExpired:
		return true
	default:
		return false
	}
}

func NewInvitationStatus(status string) (InvitationStatus, error) {
	s := InvitationStatus(status)
	if !s.IsValid() {
		return "", ErrInvalidInvitationStatus
	}
	return s, nil
}

func (s InvitationStatus) String() string {
	return string(s)
}

func (s InvitationStatus) CanTransitionTo(newStatus InvitationStatus) bool {
	if !newStatus.IsValid() {
		return false
	}

	validTransitions, ok := validInvitationStatusTransitions[s]
	if !ok {
		return false
	}

	_, canTransition := validTransitions[newStatus]
	return canTransition
}
package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

type MembershipStatus string
const (
	MembershipStatusActive MembershipStatus = "active"
	MembershipStatusSuspended MembershipStatus = "suspended"
	MembershipStatusInactive MembershipStatus = "inactive"
)

var (
	ErrInvalidMembershipStatus = customerrors.NewBadRequestError("Invalid membership status")
)

func (ms MembershipStatus) IsValid() bool {
	switch ms {
	case MembershipStatusActive, MembershipStatusSuspended, MembershipStatusInactive:
		return true
	default:
		return false
	}
}

func NewMembershipStatus(value string) (MembershipStatus, error) {
	status := MembershipStatus(value)
	if !status.IsValid() {
		return "", ErrInvalidMembershipStatus
	}
	return status, nil
}

func (ms MembershipStatus) String() string {
	return string(ms)
}
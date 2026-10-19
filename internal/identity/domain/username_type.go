package domain

import shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"

// UsernameType is re-exported from shared/domain for backward compatibility.
type UsernameType = shareddomain.UsernameType

const (
	FirstName = shareddomain.FirstName
	LastName  = shareddomain.LastName
)

var (
	ErrInvalidUsernameType = shareddomain.ErrInvalidUsernameType
	NewUsernameType        = shareddomain.NewUsernameType
)

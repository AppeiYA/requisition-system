package domain

import shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"

// Username is the identity value object for first/last names. It now lives in
// shared/domain so other contexts (e.g. invitation) can reuse the same
// validation. The alias keeps the identity package's public API stable.
type Username = shareddomain.Username

var (
	ErrUsernameTooShort = shareddomain.ErrUsernameTooShort
	ErrUsernameTooLong  = shareddomain.ErrUsernameTooLong
	ErrInvalidUsername  = shareddomain.ErrInvalidUsername
	NewUsername         = shareddomain.NewUsername
)

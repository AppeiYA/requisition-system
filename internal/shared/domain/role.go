package shareddomain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

type Role string
const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

var (
	ErrInvalidRole = customerrors.NewInternalServerError("invalid role")
)

func IsValidRole(role string) bool {
	switch Role(role) {
	case RoleAdmin, RoleUser:
		return true
	default:
		return false
	}
}

func NewRole(role string) (Role, error) {
	if !IsValidRole(role) {
		return "", ErrInvalidRole
	}
	return Role(role), nil
}

func (r Role) String() string {
	return string(r)
}
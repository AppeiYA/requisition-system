package usecase

import (
	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrRequisitionAccessDenied        = customerrors.NewForbiddenError("you are not allowed to access this requisition")
	ErrRequisitionTransitionForbidden = customerrors.NewForbiddenError("not authorized to perform this requisition status transition")
)

// actorRoleFromIdentity maps the identity role to the role stored on a status
// history row. The historical role is captured at action time and never
// recomputed from the user's current roles.
func actorRoleFromIdentity(role string) (domain.RequisitionActorRole, error) {
	switch shareddomain.Role(role) {
	case shareddomain.RoleAdmin:
		return domain.ActorRoleAdmin, nil
	case shareddomain.RoleUser:
		return domain.ActorRoleStaff, nil
	default:
		return "", domain.ErrInvalidActorRole
	}
}

func isAdminIdentityRole(role string) bool {
	return shareddomain.Role(role) == shareddomain.RoleAdmin
}

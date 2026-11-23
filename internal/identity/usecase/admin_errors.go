package usecase

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

var (
	ErrSelfUserChange           = customerrors.NewForbiddenError("cannot change your own account")
	ErrLastActiveAdmin          = customerrors.NewConflictError("cannot remove the last active admin")
	ErrReasonRequired           = customerrors.NewBadRequestError("reason is required for this status")
	ErrInvalidRoleValue         = customerrors.NewBadRequestError("invalid role")
	ErrRoleChangeRequiresActive = customerrors.NewConflictError("only active users can be promoted to admin")
	ErrInvalidUserSort          = customerrors.NewBadRequestError("invalid sort value")
	ErrUserIDRequired           = customerrors.NewBadRequestError("user id is required")
)

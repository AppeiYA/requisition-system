package usecase

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

var (
	ErrInvitationIDRequired   = customerrors.NewBadRequestError("invitation id is required")
	ErrActorIDRequired        = customerrors.NewInternalServerError("actor id is required")
	ErrInvitationCannotResend = customerrors.NewConflictError("accepted or revoked invitations cannot be resent")
	ErrInvitationCannotRevoke = customerrors.NewConflictError("accepted invitations cannot be revoked")
	ErrInvitationNotRevocable = customerrors.NewConflictError("invitation can no longer be revoked")
	ErrInvitationHasActive    = customerrors.NewConflictError("another active invitation already exists for this email")
	ErrReasonTooLong          = customerrors.NewBadRequestError("reason must be at most 500 characters")
)

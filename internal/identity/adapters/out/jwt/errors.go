package jwtadapter

import (
	"fmt"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/errors"
)

// Sentinel errors returned by the access-token service. They are
// *customerrors.CustomError values, so the transport layer can map them onto
// HTTP status codes (401 for the unauthorized ones) and callers can match them
// with errors.Is.
var (
	ErrInvalidToken            = customerrors.NewUnauthorizedError("invalid token")
	ErrExpiredToken            = customerrors.NewUnauthorizedError("token expired")
	ErrMalformedToken          = customerrors.NewUnauthorizedError("malformed token")
	ErrInvalidTokenSignature   = customerrors.NewUnauthorizedError("invalid token signature")
	ErrUnexpectedSigningMethod = customerrors.NewUnauthorizedError("unexpected signing method")
	ErrInvalidTokenModel       = customerrors.NewInternalServerError("token model must include a user id")
)

func signTokenError(err error) error {
	return customerrors.NewInternalServerError(
		fmt.Sprintf("failed to sign token: %v", err),
	)
}

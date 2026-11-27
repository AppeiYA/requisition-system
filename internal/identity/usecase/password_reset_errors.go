package usecase

import (
	"errors"
	"net/http"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	// ErrInvalidPasswordResetToken is the single generic error used for every
	// bad reset token. It deliberately gives no hint about why a token failed.
	ErrInvalidPasswordResetToken = customerrors.NewBadRequestError("invalid or expired password reset token")

	// ErrPasswordResetTokenRequired is returned when the raw token is missing
	// before any storage is touched.
	ErrPasswordResetTokenRequired = customerrors.NewBadRequestError("token is required")
)

// isNotFoundError reports whether err is the repository's not-found custom
// error. Storage not-found results are never leaked to the HTTP layer in the
// password reset flows.
func isNotFoundError(err error) bool {
	var customErr *customerrors.CustomError
	if !errors.As(err, &customErr) {
		return false
	}
	return customErr.Code == http.StatusNotFound
}

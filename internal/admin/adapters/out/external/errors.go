package adminexternal

import (
	"errors"
	"net/http"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

func isNotFound(err error) bool {
	var customErr *customerrors.CustomError
	return errors.As(err, &customErr) && customErr.Code == http.StatusNotFound
}

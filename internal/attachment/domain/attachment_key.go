package domain

import (
	"strings"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var ErrInvalidAttachmentKey = customerrors.NewBadRequestError("invalid attachment storage key")

// NewAttachmentKey derives the storage key from validated domain values. No
// client-supplied string is ever used as a path segment.
func NewAttachmentKey(ref *AttachmentRef, attachmentID string) (string, error) {
	if ref == nil {
		return "", customerrors.NewBadRequestError("attachment ref is required")
	}
	if !isValidAttachmentID(attachmentID) {
		return "", ErrInvalidAttachmentID
	}

	key := strings.Join([]string{ref.RefName().String(), ref.ID(), attachmentID}, "/")
	return key, nil
}

// Package key validates the only two object-key shapes the attachment module
// produces. Disk and object-store adapters share it so a key that is safe on
// disk is also safe in S3.
package key

import (
	"strings"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var ErrInvalidKey = customerrors.NewBadRequestError("invalid attachment storage key")

// ValidateKey accepts staging/<upl-id> and <ref-name>/<ref-id>/<att-id> only.
func ValidateKey(value string) error {
	if value == "" || strings.HasPrefix(value, "/") || strings.Contains(value, `\`) {
		return ErrInvalidKey
	}
	parts := strings.Split(value, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || strings.Contains(part, "..") {
			return ErrInvalidKey
		}
	}

	if parts[0] == domain.StagingPathPrefix {
		if len(parts) != 2 {
			return ErrInvalidKey
		}
		if _, err := domain.NewStagingUploadKey(parts[1]); err != nil {
			return ErrInvalidKey
		}
		return nil
	}

	if len(parts) != 3 {
		return ErrInvalidKey
	}
	ref, err := domain.NewAttachmentRef(domain.AttachmentRefName(parts[0]), parts[1])
	if err != nil {
		return ErrInvalidKey
	}
	if _, err := domain.NewAttachmentKey(ref, parts[2]); err != nil {
		return ErrInvalidKey
	}
	return nil
}

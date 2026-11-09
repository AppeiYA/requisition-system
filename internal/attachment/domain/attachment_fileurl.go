package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

var (
	ErrInvalidAttachmentFileUrl = customerrors.NewInternalServerError("invalid attachment file url")
)

type AttachmentFileUrl struct {
	value string
}

func NewAttachmentFileUrl(value string) (*AttachmentFileUrl, error) {
	if value == "" {
		return nil, ErrInvalidAttachmentFileUrl
	}
	return &AttachmentFileUrl{value: value}, nil
}

func (f *AttachmentFileUrl) String() string {
	return f.value
}
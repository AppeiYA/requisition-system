package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

var (
	ErrInvalidAttachmentFileName = customerrors.NewInternalServerError("invalid attachment file name")
)

type AttachmentFileName struct {
	value string 
}

func NewAttachmentFileName(value string) (*AttachmentFileName, error) {
	if value == "" {
		return nil, ErrInvalidAttachmentFileName
	}
	return &AttachmentFileName{value: value}, nil
}

func (f *AttachmentFileName) String() string {
	return f.value
}
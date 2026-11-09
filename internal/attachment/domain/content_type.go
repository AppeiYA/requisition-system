package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

type ContentType string

var (
	ErrInvalidContentType = customerrors.NewBadRequestError("invalid content type")
)

const (
	ContentTypePDF  ContentType = "application/pdf"
	ContentTypePNG  ContentType = "image/png"
	ContentTypeJPEG ContentType = "image/jpeg"
	ContentTypeGIF  ContentType = "image/gif"

	ContentTypeDOCX ContentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	ContentTypeXLSX ContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	ContentTypeCSV  ContentType = "text/csv"
)

func (ct ContentType) IsValid() bool {
	switch ct {
	case ContentTypePDF, ContentTypePNG, ContentTypeJPEG, ContentTypeGIF, ContentTypeDOCX, ContentTypeXLSX, ContentTypeCSV:
		return true
	default:
		return false
	}
}

func NewContentType(value string) (ContentType, error) {
	contentType := ContentType(value)
	if !contentType.IsValid() {
		return "", ErrInvalidContentType
	}
	return contentType, nil
}

func (ct ContentType) String() string {
	return string(ct)
}
package attachmentout

import "github.com/AppeiYA/requisition-system/internal/attachment/domain"

// ContentTypeDetector derives the stored content type from the file bytes.
// The declared type is only a hint and is never trusted on its own.
type ContentTypeDetector interface {
	Detect(head []byte, declared string) (domain.ContentType, error)
}

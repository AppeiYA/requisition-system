package attachmenthttp

import (
	"net/url"
	"strings"
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
)

// AttachmentResponse is the metadata DTO returned by every attachment
// endpoint. The URL is derived per request; no storage key or disk path is
// ever exposed.
type AttachmentResponse struct {
	ID          string    `json:"id"`
	RefName     string    `json:"ref_name"`
	RefID       string    `json:"ref_id"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	FileSize    int64     `json:"file_size"`
	Checksum    string    `json:"checksum"`
	Status      string    `json:"status"`
	UploadedBy  string    `json:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at"`
	URL         string    `json:"url"`
}

// StagedUploadResponse is the metadata DTO for a parentless upload. The
// storage path is deliberately omitted: clients only need the server-generated
// ID to reference the file during requisition creation.
type StagedUploadResponse struct {
	ID          string    `json:"id"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	FileSize    int64     `json:"file_size"`
	Checksum    string    `json:"checksum"`
	Status      string    `json:"status"`
	UploadedBy  string    `json:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func stagedUploadResponse(upload *domain.StagedUpload) *StagedUploadResponse {
	if upload == nil {
		return nil
	}

	return &StagedUploadResponse{
		ID:          upload.ID(),
		FileName:    upload.OriginalFileName().String(),
		ContentType: upload.ContentType().String(),
		FileSize:    upload.FileSize(),
		Checksum:    upload.Checksum(),
		Status:      upload.Status().String(),
		UploadedBy:  upload.UploadedBy(),
		CreatedAt:   upload.CreatedAt(),
		ExpiresAt:   upload.ExpiresAt(),
	}
}

func attachmentResponse(att *domain.Attachment, baseURL string) *AttachmentResponse {
	if att == nil {
		return nil
	}

	return &AttachmentResponse{
		ID:          att.ID(),
		RefName:     att.Ref().RefName().String(),
		RefID:       att.Ref().ID(),
		FileName:    att.FileName().String(),
		ContentType: att.ContentType().String(),
		FileSize:    att.FileSize(),
		Checksum:    att.Checksum(),
		Status:      att.Status().String(),
		UploadedBy:  att.UploadedBy(),
		CreatedAt:   att.CreatedAt(),
		URL:         attachmentContentURL(baseURL, att.ID()),
	}
}

// attachmentContentURL keeps a configured absolute base URL (for example
// http://localhost:8080/api/v1) as a path, so streamed downloads stay on the
// same host as the API request.
func attachmentContentURL(baseURL, attachmentID string) string {
	base := strings.TrimSpace(baseURL)
	if parsed, err := url.Parse(base); err == nil && parsed.Scheme != "" {
		base = parsed.Path
	}
	base = strings.TrimRight(base, "/")
	return base + "/attachments/" + url.PathEscape(attachmentID) + "/content"
}

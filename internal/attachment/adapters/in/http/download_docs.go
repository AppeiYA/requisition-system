package attachmenthttp

import (
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

// The blank assignment keeps the imported type referenced so swaggo can
// resolve it from the annotations below; it has no runtime effect.
var _ = response.Response{}

// downloadDoc documents the download handler in download.go.
//
//	@Summary		Download an attachment
//	@Description	Streams the attachment bytes with range support, or redirects to a signed URL when the storage adapter provides one.
//	@Tags			attachment
//	@Produce		application/octet-stream
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Attachment id"
//	@Success		200	{file}		file	"Attachment bytes"
//	@Failure		302	{object}	response.Response	"Redirect to a signed URL"
//	@Failure		401	{object}	response.Response	"Missing or invalid access token"
//	@Failure		403	{object}	response.Response	"Access to the parent resource denied"
//	@Failure		404	{object}	response.Response	"Attachment not found"
//	@Failure		409	{object}	response.ErrorResponse	"Attachment is not ready for download (error: attachment_not_ready)"
//	@Router			/attachments/{id}/content [get]
func downloadDoc() {}

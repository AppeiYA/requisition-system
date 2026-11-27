package attachmenthttp

import (
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

// The blank assignment keeps the imported type referenced so swaggo can
// resolve it from the annotations below; it has no runtime effect.
var _ = response.Response{}

// deleteDoc documents the delete handler in delete.go.
//
//	@Summary		Delete an attachment
//	@Description	Removes the attachment metadata and its stored object. Only the uploader or an administrator may delete it.
//	@Tags			attachment
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Attachment id"
//	@Success		200	{object}	response.Response
//	@Failure		401	{object}	response.Response	"Missing or invalid access token"
//	@Failure		403	{object}	response.Response	"Access to the parent resource denied"
//	@Failure		404	{object}	response.Response	"Attachment not found"
//	@Router			/attachments/{id} [delete]
func deleteDoc() {}

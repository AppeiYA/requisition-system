package attachmenthttp

import (
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

// The blank assignment keeps the imported type referenced so swaggo can
// resolve it from the annotations below; it has no runtime effect.
var _ = response.Response{}

// uploadDoc documents the upload handler in upload.go.
//
//	@Summary		Upload an attachment
//	@Description	Streams a file to storage and records its metadata against a requisition or requisition item. The content type is detected from the file bytes, not from the declared type.
//	@Tags			attachment
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file		formData	file	true	"File to upload"
//	@Param			ref_name	formData	string	true	"Parent type (requisition or requisition_item)"	Enums(requisition, requisition_item)
//	@Param			ref_id		formData	string	true	"Parent id"
//	@Success		201			{object}	response.Response{data=AttachmentResponse}
//	@Failure		400			{object}	response.ErrorResponse	"Invalid multipart body"
//	@Failure		401			{object}	response.Response		"Missing or invalid access token"
//	@Failure		403			{object}	response.Response		"Access to the parent resource denied"
//	@Failure		404			{object}	response.Response		"Parent resource not found"
//	@Failure		409			{object}	response.Response		"Attachment limit for the parent reached"
//	@Failure		413			{object}	response.Response		"File exceeds the configured size limit"
//	@Failure		415			{object}	response.Response		"Unsupported content type"
//	@Router			/attachments [post]
func uploadDoc() {}

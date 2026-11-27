package attachmenthttp

import (
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

// The blank assignment keeps the imported type referenced so swaggo can
// resolve it from the annotations below; it has no runtime effect.
var _ = response.Response{}

// listDoc documents the list handler in list.go.
//
//	@Summary		List attachments
//	@Description	Returns a paginated list of attachments (including their lifecycle status) for one parent resource.
//	@Tags			attachment
//	@Produce		json
//	@Security		BearerAuth
//	@Param			ref_name	query		string	true	"Parent type (requisition or requisition_item)"	Enums(requisition, requisition_item)
//	@Param			ref_id		query		string	true	"Parent id"
//	@Param			limit		query		integer	false	"Page size (1-100)"
//	@Param			offset		query		integer	false	"Page offset"
//	@Success		200			{object}	response.Response{data=[]AttachmentResponse}
//	@Failure		400			{object}	response.ErrorResponse	"Missing ref_name/ref_id or invalid pagination"
//	@Failure		401			{object}	response.Response		"Missing or invalid access token"
//	@Failure		403			{object}	response.Response		"Access to the parent resource denied"
//	@Failure		404			{object}	response.Response		"Parent resource not found"
//	@Router			/attachments [get]
func listDoc() {}

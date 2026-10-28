package requisitionhttp

import (
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

// The blank assignments keep the imported types referenced so swaggo can
// resolve them from the annotations below; they have no runtime effect.
var (
	_ = response.Response{}
	_ = RequisitionResponse{}
)

// listAllRequisitionsDoc documents the listAll handler in requisition_handler.go.
//
//	@Summary		List all requisitions
//	@Description	Returns all requisitions ordered by creation time (newest first). Requires an access token and admin role.
//	@Tags			requisition
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query	int	false	"Maximum number of requisitions to return (default 100, max 100)"
//	@Param			offset	query	int	false	"Number of requisitions to skip (default 0)"
//	@Success		200		{object}	response.Response{data=[]RequisitionResponse}
//	@Failure		400		{object}	response.Response	"Invalid pagination parameters"
//	@Failure		401		{object}	response.Response	"Missing or invalid access token"
//	@Failure		403		{object}	response.Response	"Forbidden"
//	@Router			/requisitions/all [get]
func listAllRequisitionsDoc() {}

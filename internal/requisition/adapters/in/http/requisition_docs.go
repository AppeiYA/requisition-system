package requisitionhttp

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

var (
	_ = response.Response{}
	_ = RequisitionResponse{}
)

// createRequisitionDoc documents the create handler in requisition_handler.go.
//
//	@Summary		Create a requisition
//	@Description	Creates a requisition for the authenticated user's department. The category is required; item_id is optional. needed_by must be a calendar date in YYYY-MM-DD format and cannot be before today.
//	@Tags			requisition
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateRequisitionRequest	true	"Requisition details"
//	@Success		201		{object}	response.Response{data=RequisitionResponse}
//	@Failure		400		{object}	response.ErrorResponse	"Validation error"
//	@Failure		401		{object}	response.Response		"Missing or invalid access token"
//	@Failure		404		{object}	response.Response		"Department membership, category, or item not found"
//	@Router			/requisitions [post]
func createRequisitionDoc() {}

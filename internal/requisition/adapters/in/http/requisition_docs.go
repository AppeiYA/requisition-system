package requisitionhttp

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

var (
	_ = response.Response{}
	_ = RequisitionResponse{}
)

// createRequisitionDoc documents the create handler in requisition_handler.go.
//
//	@Summary		Create a requisition
//	@Description	Creates a multi-line requisition for the authenticated user's department. Each line references either a catalogue item (item_id) or a free-filled requested item (requested_item), never both. Amounts are expressed in major currency units (naira for NGN) and stored internally as minor units. needed_by must be a calendar date in YYYY-MM-DD format and must be in the future.
//	@Tags			requisition
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateRequisitionRequest	true	"Requisition details"
//	@Success		201		{object}	response.Response{data=RequisitionResponse}
//	@Failure		400		{object}	response.ErrorResponse	"Validation error"
//	@Failure		401		{object}	response.Response		"Missing or invalid access token"
//	@Failure		403		{object}	response.Response		"Department membership is inactive or belongs to another department"
//	@Failure		404		{object}	response.Response		"Department membership, category, or item not found"
//	@Router			/requisitions [post]
func createRequisitionDoc() {}

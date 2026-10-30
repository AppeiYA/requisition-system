package requisitionhttp

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

// The blank assignments keep the imported types referenced so swaggo can
// resolve them from the annotations below; they have no runtime effect.
var (
	_ = response.Response{}
	_ = RequisitionResponse{}
)

// getRequisitionDoc documents the getByID handler in requisition_handler.go.
//
//	@Summary		Get a requisition
//	@Description	Returns a single requisition with its lines, derived totals, and status history. Requesters may only fetch their own requisitions; administrators may fetch any. Amounts are expressed in major currency units (naira for NGN) while stored internally as minor units.
//	@Tags			requisition
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Requisition ID"
//	@Success		200	{object}	response.Response{data=RequisitionResponse}
//	@Failure		400	{object}	response.Response	"Invalid requisition ID"
//	@Failure		401	{object}	response.Response	"Missing or invalid access token"
//	@Failure		403	{object}	response.Response	"Requisition belongs to another user"
//	@Failure		404	{object}	response.Response	"Requisition not found"
//	@Router			/requisitions/{id} [get]
func getRequisitionDoc() {}

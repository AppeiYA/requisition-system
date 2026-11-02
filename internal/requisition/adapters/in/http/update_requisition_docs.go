package requisitionhttp

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

var _ = response.Response{}

// updateRequisitionDoc documents the pending requisition edit handler.
//
//	@Summary		Update a pending requisition
//	@Description	Replaces the header and line values of a pending requisition owned by the authenticated requester. Amounts are expressed in major currency units (naira for NGN) and stored internally as minor units.
//	@Tags			requisition
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string	true	"Requisition ID"
//	@Param			request	body		CreateRequisitionRequest	true	"Updated requisition details"
//	@Success		200		{object}	response.Response{data=RequisitionResponse}
//	@Failure		400		{object}	response.ErrorResponse	"Validation error"
//	@Failure		401		{object}	response.Response		"Missing or invalid access token"
//	@Failure		403		{object}	response.Response		"Requester does not own the requisition"
//	@Failure		409		{object}	response.Response		"Requisition is no longer pending"
//	@Router			/requisitions/{id} [patch]
func updateRequisitionDoc() {}

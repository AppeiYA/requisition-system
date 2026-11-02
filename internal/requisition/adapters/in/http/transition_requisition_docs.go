package requisitionhttp

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

// The blank assignments keep the imported types referenced so swaggo can
// resolve them from the annotations below; they have no runtime effect.
var (
	_ = response.Response{}
	_ = RequisitionResponse{}
)

// transitionRequisitionDoc documents the transitionStatus handler in
// requisition_handler.go.
//
//	@Summary		Transition a requisition status
//	@Description	Applies a status transition and records it in the append-only status history. A reason is required for rejected and cancelled. Requesters may only cancel their own pending requisition; approval, rejection, fulfillment, and completion require an administrator. Returns the complete updated requisition. Amounts are expressed in major currency units (naira for NGN) while stored internally as minor units.
//	@Tags			requisition
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string								true	"Requisition ID"
//	@Param			request	body	TransitionRequisitionStatusRequest	true	"Target status and optional reason"
//	@Success		200		{object}	response.Response{data=RequisitionResponse}
//	@Failure		400		{object}	response.ErrorResponse	"Invalid status, missing reason, or invalid transition"
//	@Failure		401		{object}	response.Response		"Missing or invalid access token"
//	@Failure		403		{object}	response.Response		"Actor is not authorized for this transition"
//	@Failure		404		{object}	response.Response		"Requisition not found"
//	@Failure		409		{object}	response.Response		"Concurrent status change"
//	@Router			/requisitions/{id}/status [post]
func transitionRequisitionDoc() {}

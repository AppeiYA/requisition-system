package item

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

var _ = response.Response{}

// promoteRequestedItemDoc documents the promoteRequested handler.
//
//	@Summary		Promote a requested item
//	@Description	Creates a catalogue item from a requested item and records the administrator review.
//	@Tags			item
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string							true	"Requested item ID"
//	@Param			request	body		PromoteRequestedItemRequest	true	"Promotion details"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.Response	"Validation or invalid review state"
//	@Failure		401		{object}	response.Response	"Missing or invalid access token"
//	@Failure		403		{object}	response.Response	"Administrator access required"
//	@Failure		404		{object}	response.Response	"Requested item not found"
//	@Failure		409		{object}	response.Response	"Requested item already reviewed or catalogue duplicate"
//	@Router			/items/requested/{id}/promote [post]
func promoteRequestedItemDoc() {}

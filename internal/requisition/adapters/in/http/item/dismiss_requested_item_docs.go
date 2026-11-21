package item

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

var _ = response.Response{}

// dismissRequestedItemDoc documents the dismissRequested handler.
//
//	@Summary		Dismiss a requested item
//	@Description	Marks a requested item as dismissed and records the administrator review.
//	@Tags			item
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Requested item ID"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response	"Invalid review state"
//	@Failure		401	{object}	response.Response	"Missing or invalid access token"
//	@Failure		403	{object}	response.Response	"Administrator access required"
//	@Failure		404	{object}	response.Response	"Requested item not found"
//	@Failure		409	{object}	response.Response	"Requested item already reviewed"
//	@Router			/items/requested/{id}/dismiss [post]
func dismissRequestedItemDoc() {}

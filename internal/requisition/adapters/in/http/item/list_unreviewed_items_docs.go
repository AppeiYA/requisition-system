package item

import (
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

var _ = response.Response{}

// listUnreviewedItemsDoc documents the listUnreviewed handler.
//
//	@Summary		List unreviewed requested items
//	@Description	Returns free-filled requested items awaiting administrator review.
//	@Tags			item
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query	int	false	"Maximum number of items to return (default 100, max 100)"
//	@Param			offset	query	int	false	"Number of items to skip (default 0)"
//	@Success		200		{object}	response.Response{data=[]RequestedItemResponse}
//	@Failure		400		{object}	response.Response	"Invalid pagination parameters"
//	@Failure		401		{object}	response.Response	"Missing or invalid access token"
//	@Failure		403		{object}	response.Response	"Administrator access required"
//	@Router			/items/requested/unreviewed [get]
func listUnreviewedItemsDoc() {}

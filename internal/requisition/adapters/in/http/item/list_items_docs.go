package item

import (
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

// The blank assignments keep the imported types referenced so swaggo can
// resolve them from the annotations below; they have no runtime effect.
var (
	_ = response.Response{}
	_ = ItemResponse{}
)

// listItemsDoc documents the list handler in item_handler.go.
//
//	@Summary		List items
//	@Description	Returns items ordered by name. Requires an access token.
//	@Tags			item
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query	int	false	"Maximum number of items to return (default 100, max 100)"
//	@Param			offset	query	int	false	"Number of items to skip (default 0)"
//	@Success		200		{object}	response.Response{data=[]ItemResponse}
//	@Failure		400		{object}	response.Response	"Invalid pagination parameters"
//	@Failure		401		{object}	response.Response	"Missing or invalid access token"
//	@Router			/items [get]
func listItemsDoc() {}

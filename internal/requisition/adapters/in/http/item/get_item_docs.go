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

// getItemDoc documents the getByID handler in item_handler.go.
//
//	@Summary		Get an item
//	@Description	Returns a single item by ID. Requires an access token.
//	@Tags			item
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Item ID"
//	@Success		200	{object}	response.Response{data=ItemResponse}
//	@Failure		401	{object}	response.Response	"Missing or invalid access token"
//	@Failure		404	{object}	response.Response	"Item not found"
//	@Router			/items/{id} [get]
func getItemDoc() {}

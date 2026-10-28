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

// createItemDoc documents the create handler in item_handler.go.
//
//	@Summary		Create an item
//	@Description	Creates an item in an existing category. Requires an admin access token.
//	@Tags			item
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateItemRequest	true	"Item details"
//	@Success		201		{object}	response.Response{data=ItemResponse}
//	@Failure		400		{object}	response.ErrorResponse	"Validation error"
//	@Failure		401		{object}	response.Response		"Missing or invalid access token"
//	@Failure		403		{object}	response.Response		"Admin role required"
//	@Failure		404		{object}	response.Response		"Category not found"
//	@Failure		409		{object}	response.Response		"Item name already exists in this category"
//	@Router			/items [post]
func createItemDoc() {}

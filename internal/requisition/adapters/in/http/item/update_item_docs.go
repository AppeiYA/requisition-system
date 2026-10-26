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

// updateItemDoc documents the update handler in item_handler.go.
//
//	@Summary		Update an item
//	@Description	Partially updates an item's name, description, category and unit cost. Every field is optional, but at least one field must be provided. Requires an admin access token.
//	@Tags			item
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string				true	"Item ID"
//	@Param			request	body	UpdateItemRequest	true	"Fields to update"
//	@Success		200		{object}	response.Response{data=ItemResponse}
//	@Failure		400		{object}	response.ErrorResponse	"Validation error or no fields provided"
//	@Failure		401		{object}	response.Response		"Missing or invalid access token"
//	@Failure		403		{object}	response.Response		"Admin role required"
//	@Failure		404		{object}	response.Response		"Item or category not found"
//	@Failure		409		{object}	response.Response		"Item name already exists in this category"
//	@Router			/items/{id} [patch]
func updateItemDoc() {}

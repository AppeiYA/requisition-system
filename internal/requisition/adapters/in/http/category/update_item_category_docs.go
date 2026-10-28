package category

import (
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

// The blank assignments keep the imported types referenced so swaggo can
// resolve them from the annotations below; they have no runtime effect.
var (
	_ = response.Response{}
	_ = ItemCategoryResponse{}
)

// updateItemCategoryDoc documents the update handler in handler.go.
//
//	@Summary		Update an item category
//	@Description	Partially updates an item category's name and description. Every field is optional, but at least one of name or description must be provided. Requires an admin access token.
//	@Tags			category
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string						true	"Category ID"
//	@Param			request	body	UpdateItemCategoryRequest	true	"Fields to update"
//	@Success		200		{object}	response.Response{data=ItemCategoryResponse}
//	@Failure		400		{object}	response.ErrorResponse	"Validation error or no name/description provided"
//	@Failure		401		{object}	response.Response		"Missing or invalid access token"
//	@Failure		403		{object}	response.Response		"Admin role required"
//	@Failure		404		{object}	response.Response		"Category not found"
//	@Failure		409		{object}	response.Response		"Category name already exists"
//	@Router			/categories/{id} [patch]
func updateItemCategoryDoc() {}

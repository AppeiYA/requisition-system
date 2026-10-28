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

// getItemCategoryDoc documents the getByID handler in handler.go.
//
//	@Summary		Get an item category
//	@Description	Returns a single item category by ID. Requires an access token.
//	@Tags			category
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Category ID"
//	@Success		200	{object}	response.Response{data=ItemCategoryResponse}
//	@Failure		401	{object}	response.Response	"Missing or invalid access token"
//	@Failure		404	{object}	response.Response	"Category not found"
//	@Router			/categories/{id} [get]
func getItemCategoryDoc() {}

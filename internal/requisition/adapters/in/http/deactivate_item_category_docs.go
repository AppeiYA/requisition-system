package requisitionhttp

import (
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

// The blank assignments keep the imported types referenced so swaggo can
// resolve them from the annotations below; they have no runtime effect.
var (
	_ = response.Response{}
	_ = ItemCategoryResponse{}
)

// deactivateItemCategoryDoc documents the deactivate handler in handler.go.
//
//	@Summary		Deactivate an item category
//	@Description	Marks an item category as inactive. Requires an admin access token.
//	@Tags			category
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Category ID"
//	@Success		200	{object}	response.Response{data=ItemCategoryResponse}
//	@Failure		401	{object}	response.Response	"Missing or invalid access token"
//	@Failure		403	{object}	response.Response	"Admin role required"
//	@Failure		404	{object}	response.Response	"Category not found"
//	@Router			/categories/{id}/deactivate [patch]
func deactivateItemCategoryDoc() {}

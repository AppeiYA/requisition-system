package category

import (
	itemhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http/item"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

// The blank assignments keep the imported types referenced so swaggo can
// resolve them from the annotations below; they have no runtime effect.
var (
	_ = response.Response{}
	_ = itemhttp.ItemResponse{}
)

// listItemsInCategoryDoc documents the listItems handler in handler.go.
//
//	@Summary		List items in a category
//	@Description	Returns items belonging to the specified category. Requires an access token.
//	@Tags			category
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Category ID"
//	@Success		200	{object}	response.Response{data=[]item.ItemResponse}
//	@Failure		400	{object}	response.Response	"Invalid category ID"
//	@Failure		401	{object}	response.Response	"Missing or invalid access token"
//	@Failure		404	{object}	response.Response	"Category not found"
//	@Router			/categories/{id}/items [get]
func listItemsInCategoryDoc() {}

// listItemsInCategoryDocWithPagination notes pagination.
//	@Param			limit	query	int	false	"Maximum number of items to return (default 100, max 100)"
//	@Param			offset	query	int	false	"Number of items to skip (default 0)"
func listItemsInCategoryDocWithPagination() {}

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

// listItemCategoriesDoc documents the list handler in handler.go.
//
//	@Summary		List item categories
//	@Description	Returns item categories ordered by name. Requires an access token.
//	@Tags			category
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query	int	false	"Maximum number of categories to return (default 100, max 100)"
//	@Param			offset	query	int	false	"Number of categories to skip (default 0)"
//	@Success		200		{object}	response.Response{data=[]ItemCategoryResponse}
//	@Failure		400		{object}	response.Response	"Invalid pagination parameters"
//	@Failure		401		{object}	response.Response	"Missing or invalid access token"
//	@Router			/categories [get]
func listItemCategoriesDoc() {}

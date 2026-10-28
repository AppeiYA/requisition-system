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

// createItemCategoryDoc documents the create handler in handler.go.
//
//	@Summary		Create an item category
//	@Description	Creates an item category. Requires an admin access token.
//	@Tags			category
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateItemCategoryRequest	true	"Category details"
//	@Success		201		{object}	response.Response{data=ItemCategoryResponse}
//	@Failure		400		{object}	response.ErrorResponse	"Validation error"
//	@Failure		401		{object}	response.Response		"Missing or invalid access token"
//	@Failure		403		{object}	response.Response		"Admin role required"
//	@Failure		409		{object}	response.Response		"Category name already exists"
//	@Router			/categories [post]
func createItemCategoryDoc() {}

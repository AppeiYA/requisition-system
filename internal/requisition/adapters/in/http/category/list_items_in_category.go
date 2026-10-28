package category

import (
	"net/http"

	"github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http/item"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *ItemCategoryHandler) listItems(c *gin.Context) {
	items, err := h.itemCategoryPortIn.ListItemsInCategoryUsecase.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	result := make([]*item.ItemResponse, 0, len(items))
	for _, it := range items {
		result = append(result, item.ItemResponseFromDomain(it))
	}

	response.Success(c, "items fetched successfully", http.StatusOK, result)
}

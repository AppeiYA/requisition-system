package item

import (
	"net/http"

	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/pagination"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *ItemHandler) list(c *gin.Context) {
	q, ok := pagination.Parse(c)
	if !ok {
		return
	}

	items, err := h.itemPortIn.ListItemsUsecase.Execute(c.Request.Context(), itemin.ListItemsQuery{
		Limit:  q.Limit,
		Offset: q.Offset,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	result := make([]*ItemResponse, 0, len(items))
	for _, item := range items {
		result = append(result, ItemResponseFromDomain(item))
	}

	response.SuccessWithPagination(c, "items fetched successfully", http.StatusOK, result, response.NewPaginationMeta(q.Limit, q.Offset))
}

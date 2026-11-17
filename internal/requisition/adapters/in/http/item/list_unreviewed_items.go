package item

import (
	"net/http"

	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/pagination"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *ItemHandler) listUnreviewed(c *gin.Context) {
	q, ok := pagination.Parse(c)
	if !ok {
		return
	}

	items, err := h.itemPortIn.ListUnreviewedItemsUsecase.Execute(c.Request.Context(), itemin.ListItemsQuery{
		Limit: q.Limit, Offset: q.Offset,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	result := make([]*RequestedItemResponse, 0, len(items))
	for _, item := range items {
		result = append(result, RequestedItemResponseFromDomain(item))
	}
	response.SuccessWithPagination(c, "unreviewed requested items fetched successfully", http.StatusOK, result, response.NewPaginationMeta(q.Limit, q.Offset))
}

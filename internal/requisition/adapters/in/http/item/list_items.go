package item

import (
	"net/http"
	"strconv"

	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

const (
	defaultItemLimit = 100
	maxItemLimit     = 100
)

func (h *ItemHandler) list(c *gin.Context) {
	limit, offset, ok := itemPagination(c)
	if !ok {
		return
	}

	items, err := h.itemPortIn.ListItemsUsecase.Execute(c.Request.Context(), itemin.ListItemsQuery{
		Limit:  limit,
		Offset: offset,
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

	response.Success(c, "items fetched successfully", http.StatusOK, result)
}

func itemPagination(c *gin.Context) (int, int, bool) {
	limit := defaultItemLimit
	if raw := c.Query("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > maxItemLimit {
			response.Error(c, "invalid limit: must be an integer between 1 and 100", http.StatusBadRequest, nil)
			return 0, 0, false
		}
		limit = value
	}

	offset := 0
	if raw := c.Query("offset"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			response.Error(c, "invalid offset: must be a non-negative integer", http.StatusBadRequest, nil)
			return 0, 0, false
		}
		offset = value
	}

	return limit, offset, true
}

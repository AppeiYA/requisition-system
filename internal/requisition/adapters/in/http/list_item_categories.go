package requisitionhttp

import (
	"net/http"
	"strconv"

	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

const (
	defaultItemCategoryLimit = 100
	maxItemCategoryLimit     = 100
)

func (h *ItemCategoryHandler) list(c *gin.Context) {
	limit, offset, ok := itemCategoryPagination(c)
	if !ok {
		return
	}

	categories, err := h.itemCategoryPortIn.ListItemCategoriesUsecase.Execute(c.Request.Context(), requisitionin.ListItemCategoriesQuery{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	result := make([]*ItemCategoryResponse, 0, len(categories))
	for _, category := range categories {
		result = append(result, ItemCategoryResponseFromDomain(category))
	}

	response.Success(c, "item categories fetched successfully", http.StatusOK, result)
}

func itemCategoryPagination(c *gin.Context) (int, int, bool) {
	limit := defaultItemCategoryLimit
	if raw := c.Query("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > maxItemCategoryLimit {
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

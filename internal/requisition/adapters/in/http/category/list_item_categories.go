package category

import (
	"net/http"

	categoryin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/category"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/pagination"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *ItemCategoryHandler) list(c *gin.Context) {
	q, ok := pagination.Parse(c)
	if !ok {
		return
	}

	categories, err := h.itemCategoryPortIn.ListItemCategoriesUsecase.Execute(c.Request.Context(), categoryin.ListItemCategoriesQuery{
		Limit:  q.Limit,
		Offset: q.Offset,
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

	response.SuccessWithPagination(c, "item categories fetched successfully", http.StatusOK, result, response.NewPaginationMeta(q.Limit, q.Offset))
}

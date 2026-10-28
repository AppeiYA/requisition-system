package category

import (
	"net/http"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *ItemCategoryHandler) getByID(c *gin.Context) {
	category, err := h.itemCategoryPortIn.GetItemCategoryUsecase.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "item category fetched successfully", http.StatusOK, ItemCategoryResponseFromDomain(category))
}

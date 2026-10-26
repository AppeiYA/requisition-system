package category

import (
	"net/http"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *ItemCategoryHandler) deactivate(c *gin.Context) {
	category, err := h.itemCategoryPortIn.DeactivateItemCategoryUsecase.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "item category deactivated", http.StatusOK, ItemCategoryResponseFromDomain(category))
}

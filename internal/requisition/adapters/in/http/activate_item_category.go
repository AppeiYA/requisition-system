package requisitionhttp

import (
	"net/http"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *ItemCategoryHandler) activate(c *gin.Context) {
	category, err := h.itemCategoryPortIn.ActivateItemCategoryUsecase.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "item category activated", http.StatusOK, ItemCategoryResponseFromDomain(category))
}

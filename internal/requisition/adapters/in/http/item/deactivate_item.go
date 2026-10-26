package item

import (
	"net/http"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *ItemHandler) deactivate(c *gin.Context) {
	item, err := h.itemPortIn.DeactivateItemUsecase.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "item deactivated", http.StatusOK, ItemResponseFromDomain(item))
}

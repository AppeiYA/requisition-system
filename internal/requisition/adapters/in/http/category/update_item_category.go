package category

import (
	"net/http"

	categoryin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/category"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *ItemCategoryHandler) update(c *gin.Context) {
	var req UpdateItemCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, response.ValidationErrors(err))
		c.Abort()
		return
	}

	category, err := h.itemCategoryPortIn.UpdateItemCategoryUsecase.Execute(c.Request.Context(), categoryin.UpdateItemCategoryCommand{
		ID:          c.Param("id"),
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "item category updated", http.StatusOK, ItemCategoryResponseFromDomain(category))
}

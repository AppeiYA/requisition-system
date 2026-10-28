package category

import (
	"net/http"

	categoryin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/category"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *ItemCategoryHandler) create(c *gin.Context) {
	var req CreateItemCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, response.ValidationErrors(err))
		c.Abort()
		return
	}

	category, err := h.itemCategoryPortIn.CreateItemCategoryUsecase.Execute(c.Request.Context(), categoryin.CreateItemCategoryCommand{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "item category created", http.StatusCreated, ItemCategoryResponseFromDomain(category))
}

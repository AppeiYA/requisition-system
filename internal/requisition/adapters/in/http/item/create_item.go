package item

import (
	"net/http"

	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *ItemHandler) create(c *gin.Context) {
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, response.ValidationErrors(err))
		c.Abort()
		return
	}

	item, err := h.itemPortIn.CreateItemUsecase.Execute(c.Request.Context(), itemin.CreateItemCommand{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		UnitCost:    req.UnitCost,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "item created", http.StatusCreated, ItemResponseFromDomain(item))
}

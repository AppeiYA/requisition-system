package item

import (
	"github.com/gin-gonic/gin"

	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
)

type ItemHandler struct {
	itemPortIn *itemin.ItemPortIn
}

func NewItemHandler(
	itemPortIn *itemin.ItemPortIn,
) *ItemHandler {
	return &ItemHandler{
		itemPortIn: itemPortIn,
	}
}

func (h *ItemHandler) Register(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
	adminOnlyMiddleware gin.HandlerFunc,
) {
	items := router.Group("/items")
	items.GET("", authMiddleware, h.list)
	items.GET("/:id", authMiddleware, h.getByID)
	items.POST("", authMiddleware, adminOnlyMiddleware, h.create)
	items.PATCH("/:id", authMiddleware, adminOnlyMiddleware, h.update)
	items.PATCH("/:id/activate", authMiddleware, adminOnlyMiddleware, h.activate)
	items.PATCH("/:id/deactivate", authMiddleware, adminOnlyMiddleware, h.deactivate)
}

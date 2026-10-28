package category

import (
	"github.com/gin-gonic/gin"

	categoryin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/category"
)

type ItemCategoryHandler struct {
	itemCategoryPortIn *categoryin.ItemCategoryPortIn
}

func NewItemCategoryHandler(
	itemCategoryPortIn *categoryin.ItemCategoryPortIn,
) *ItemCategoryHandler {
	return &ItemCategoryHandler{
		itemCategoryPortIn: itemCategoryPortIn,
	}
}

func (h *ItemCategoryHandler) Register(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
	adminOnlyMiddleware gin.HandlerFunc,
) {
	categories := router.Group("/categories")
	categories.GET("", authMiddleware, h.list)
	categories.GET("/:id", authMiddleware, h.getByID)
	categories.GET("/:id/items", authMiddleware, h.listItems)
	categories.POST("", authMiddleware, adminOnlyMiddleware, h.create)
	categories.PATCH("/:id", authMiddleware, adminOnlyMiddleware, h.update)
	categories.PATCH("/:id/activate", authMiddleware, adminOnlyMiddleware, h.activate)
	categories.PATCH("/:id/deactivate", authMiddleware, adminOnlyMiddleware, h.deactivate)
}

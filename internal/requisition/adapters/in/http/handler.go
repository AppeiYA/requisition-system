package requisitionhttp

import (
	"github.com/gin-gonic/gin"

	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
)

type ItemCategoryHandler struct {
	itemCategoryPortIn *requisitionin.ItemCategoryPortIn
}

func NewItemCategoryHandler(
	itemCategoryPortIn *requisitionin.ItemCategoryPortIn,
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
	categories.POST("", authMiddleware, adminOnlyMiddleware, h.create)
	categories.PATCH("/:id", authMiddleware, adminOnlyMiddleware, h.update)
	categories.PATCH("/:id/activate", authMiddleware, adminOnlyMiddleware, h.activate)
	categories.PATCH("/:id/deactivate", authMiddleware, adminOnlyMiddleware, h.deactivate)
}

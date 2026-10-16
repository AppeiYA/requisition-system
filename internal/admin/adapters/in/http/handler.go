package adminhttp

import (
	"github.com/gin-gonic/gin"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
)

type Handler struct {
	adminPortIn *adminin.AdminPortIn
}

func NewHandler(
	adminPortIn *adminin.AdminPortIn,
) *Handler {
	return &Handler{
		adminPortIn: adminPortIn,
	}
}

func (h *Handler) Register(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := router.Group("/admin")
	admin.POST("/invitations", authMiddleware, h.invite)
}

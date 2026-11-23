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

func (h *Handler) Register(router *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	admin := router.Group("/admin")
	// Metadata goes first so the authenticated user and every handler below
	// see the request id in the context.
	admin.Use(RequestMetadata())
	admin.Use(middlewares...)
	admin.POST("/invitations", h.invite)
	admin.GET("/invitations", h.listInvitations)
	admin.GET("/invitations/:id", h.getInvitation)
	admin.POST("/invitations/:id/resend", h.resendInvitation)
	admin.POST("/invitations/:id/revoke", h.revokeInvitation)
	admin.GET("/users", h.listUsers)
	admin.GET("/users/:id", h.getUser)
	admin.PATCH("/users/:id/status", h.updateUserStatus)
	admin.PATCH("/users/:id/role", h.updateUserRole)
	admin.POST("/users/:id/sessions/revoke", h.revokeUserSessions)
}

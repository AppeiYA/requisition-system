package identityhttp

import (
	"github.com/gin-gonic/gin"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
)

type Handler struct {
	identityPortIn *identityin.IdentityPortIn
	authMiddleware *AuthMiddleware
}

func NewHandler(identityPortIn *identityin.IdentityPortIn, authMiddleware *AuthMiddleware) *Handler {
	return &Handler{
		identityPortIn: identityPortIn,
		authMiddleware: authMiddleware,
	}
}


func (h *Handler) Register(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	auth.POST("/login", h.login)
	auth.POST("/refresh", h.refresh)
	auth.POST("/logout", h.logout)

	auth.GET("/me", h.authMiddleware.Handle(), h.getMe)
}

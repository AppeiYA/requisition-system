package app

import (
	"net/http"

	adminhttp "github.com/AppeiYA/requisition-system/internal/admin/adapters/in/http"
	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	invitationhttp "github.com/AppeiYA/requisition-system/internal/invitation/adapters/in/http"
	"github.com/gin-gonic/gin"
)

func NewRouter(
	router *gin.RouterGroup,
	identityHandler *identityhttp.Handler,
	invitationHandler *invitationhttp.InvitationHandler,
	adminHandler *adminhttp.Handler,
	accessTokenService identityout.AccessTokenService,
) {
	v1 := router.Group("/v1")

	v1.GET("/active", func(c *gin.Context) {
		c.JSON(http.StatusOK, "This works")
	})

	identityHandler.Register(v1)
	invitationHandler.Register(v1)
	adminHandler.Register(v1, identityhttp.AuthMiddleware(accessTokenService))
}

package app

import (
	"net/http"

	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	"github.com/gin-gonic/gin"
)

func NewRouter(router *gin.RouterGroup, identityHandler *identityhttp.Handler) {
	v1 := router.Group("/v1")

	v1.GET("/active", func(c *gin.Context) {
		c.JSON(http.StatusOK, "This works")
	})

	identityHandler.Register(v1)
}

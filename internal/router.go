package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(router *gin.RouterGroup) {
	v1 := router.Group("/v1")

	v1.GET("/active", func(c *gin.Context) {
		c.JSON(http.StatusOK, "This works")
	})
}
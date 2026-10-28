package app

import (
	"fmt"
	"net/http"

	"github.com/AppeiYA/requisition-system/docs"
	adminhttp "github.com/AppeiYA/requisition-system/internal/admin/adapters/in/http"
	departmenthttp "github.com/AppeiYA/requisition-system/internal/department/adapters/in/http"
	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	invitationhttp "github.com/AppeiYA/requisition-system/internal/invitation/adapters/in/http"
	requisitionhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http"
	categoryhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http/category"
	itemhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http/item"
	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func NewRouter(
	router *gin.RouterGroup,
	identityHandler *identityhttp.Handler,
	invitationHandler *invitationhttp.InvitationHandler,
	adminHandler *adminhttp.Handler,
	departmentHandler *departmenthttp.DepartmentHandler,
	itemCategoryHandler *categoryhttp.ItemCategoryHandler,
	itemHandler *itemhttp.ItemHandler,
	requisitionHandler *requisitionhttp.RequisitionHandler,
	authMiddleware *identityhttp.AuthMiddleware,
) error {
	scalarHTML, err := renderScalarReference()
	if err != nil {
		return err
	}

	v1 := router.Group("/v1")

	v1.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(scalarHTML))
	})

	v1.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", docs.OpenAPISpecYAML)
	})

	v1.GET("/active", func(c *gin.Context) {
		c.JSON(http.StatusOK, "This works")
	})

	identityHandler.Register(v1)
	invitationHandler.Register(v1)
	adminHandler.Register(v1, authMiddleware.Handle(), authMiddleware.RequireRole("admin"))
	departmentHandler.Register(v1, authMiddleware.Handle())
	itemCategoryHandler.Register(v1, authMiddleware.Handle(), authMiddleware.RequireRole("admin"))
	itemHandler.Register(v1, authMiddleware.Handle(), authMiddleware.RequireRole("admin"))
	requisitionHandler.Register(v1, authMiddleware.Handle(), authMiddleware.RequireRole("admin"))

	return nil
}

const maxRequestBodyBytes = 1 << 20

func requestBodyLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
		c.Next()
	}
}

// renderScalarReference renders the Scalar API reference page from the
// embedded OpenAPI 3 specification.
func renderScalarReference() (string, error) {
	var spec map[string]any
	if err := yaml.Unmarshal(docs.OpenAPISpecYAML, &spec); err != nil {
		return "", fmt.Errorf("parse embedded OpenAPI spec: %w", err)
	}

	html, err := scalar.ApiReferenceHTML(&scalar.Options{
		SpecContent: spec,
		CustomOptions: scalar.CustomOptions{
			PageTitle: "Requisition System API",
		},
	})
	if err != nil {
		return "", fmt.Errorf("render Scalar API reference: %w", err)
	}

	return html, nil
}

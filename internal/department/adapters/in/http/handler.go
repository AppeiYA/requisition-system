package departmenthttp

import (
	"net/http"

	"github.com/gin-gonic/gin"

	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

type DepartmentHandler struct {
	createDepartmentUC departmentin.CreateDepartmentUsecaseInt
	getDepartmentUC    departmentin.GetDepartmentByIDUsecaseInt
}

func NewDepartmentHandler(
	createDepartmentUC departmentin.CreateDepartmentUsecaseInt,
	getDepartmentUC departmentin.GetDepartmentByIDUsecaseInt,
) *DepartmentHandler {
	return &DepartmentHandler{
		createDepartmentUC: createDepartmentUC,
		getDepartmentUC:    getDepartmentUC,
	}
}

func (h *DepartmentHandler) create(c *gin.Context) {
	var req CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, response.ValidationErrors(err))
		c.Abort()
		return
	}

	res, err := h.createDepartmentUC.Execute(c.Request.Context(), departmentin.CreateDepartmentCommand{
		Name:        req.Name,
		Description: req.Description,
		Code:        req.Code,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "department created", http.StatusCreated, res)
}

func (h *DepartmentHandler) getByID(c *gin.Context) {
	res, err := h.getDepartmentUC.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "department retrieved", http.StatusOK, res)
}

func (h *DepartmentHandler) Register(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := router.Group("/admin")
	admin.POST("/departments", authMiddleware, h.create)
	admin.GET("/departments/:id", authMiddleware, h.getByID)
}

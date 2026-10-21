package departmenthttp

import (
	"github.com/gin-gonic/gin"

	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
)

type DepartmentHandler struct {
	departmentPortIn *departmentin.DepartmentPortIn
}

func NewDepartmentHandler(
	departmentPortIn *departmentin.DepartmentPortIn,
) *DepartmentHandler {
	return &DepartmentHandler{
		departmentPortIn: departmentPortIn,
	}
}

func (h *DepartmentHandler) Register(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	dept := router.Group("/departments")
	dept.POST("", authMiddleware, h.create)
	dept.GET("", authMiddleware, h.getDepartments)
	dept.GET("/:id", authMiddleware, h.getByID)
}

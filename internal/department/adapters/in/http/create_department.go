package departmenthttp

import (
	"net/http"

	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *DepartmentHandler) create(c *gin.Context) {
	var req CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, response.ValidationErrors(err))
		c.Abort()
		return
	}

	res, err := h.departmentPortIn.CreateDepartmentUsecase.Execute(c.Request.Context(), departmentin.CreateDepartmentCommand{
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

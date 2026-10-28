package departmenthttp

import (
	"net/http"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *DepartmentHandler) getByID(c *gin.Context) {
	res, err := h.departmentPortIn.GetDepartmentByIDUsecase.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "department retrieved", http.StatusOK, res)
}

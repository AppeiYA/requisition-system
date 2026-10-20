package departmenthttp

import (
	"net/http"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *DepartmentHandler) getDepartments(c *gin.Context) {
	res, err := h.departmentPortIn.GetDepartmentsUsecase.Execute(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	result := make([]*DepartmentResponse, 0, len(res))
	for _, dept := range res {
		result = append(result, DepartmentResponseFromDomain(dept))
	}

	response.Success(c, "departments fetched successfully", http.StatusOK, result)
}

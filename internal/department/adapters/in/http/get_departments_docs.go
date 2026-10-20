package departmenthttp

import (
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

// The blank assignment keeps the imported type referenced so swaggo can
// resolve it from the annotations below; it has no runtime effect.
var _ = response.Response{}

// getDepartmentsDoc documents the getDepartments handler in handler.go.
//
//	@Summary		List departments
//	@Description	Returns every department, ordered by name.
//	@Tags			department
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response{data=[]DepartmentResponse}
//	@Failure		401	{object}	response.Response	"Missing or invalid access token"
//	@Router			/departments [get]
func getDepartmentsDoc() {}

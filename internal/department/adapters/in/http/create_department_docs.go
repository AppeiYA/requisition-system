package departmenthttp

import (
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

// The blank assignments keep the imported types referenced so swaggo can
// resolve them from the annotations below; they have no runtime effect.
var (
	_ = response.Response{}
	_ = departmentin.CreateDepartmentResult{}
)

// createDepartmentDoc documents the create handler in handler.go.
//
//	@Summary		Create a department
//	@Description	Creates a new department with the given name, description and code.
//	@Tags			department
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateDepartmentRequest	true	"Department details"
//	@Success		201		{object}	response.Response{data=departmentin.CreateDepartmentResult}
//	@Failure		400		{object}	response.ErrorResponse	"Validation error"
//	@Failure		401		{object}	response.Response		"Missing or invalid access token"
//	@Failure		409		{object}	response.Response		"Department already exists"
//	@Router			/departments [post]
func createDepartmentDoc() {}

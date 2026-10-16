package departmenthttp

import (
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

// The blank assignments keep the imported types referenced so swaggo can
// resolve them from the annotations below; they have no runtime effect.
var (
	_ = response.Response{}
	_ = departmentin.DepartmentResult{}
)

// getDepartmentDoc documents the getByID handler in handler.go.
//
//	@Summary		Get a department
//	@Description	Returns a single department by its id.
//	@Tags			department
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Department id"
//	@Success		200	{object}	response.Response{data=departmentin.DepartmentResult}
//	@Failure		401	{object}	response.Response	"Missing or invalid access token"
//	@Failure		404	{object}	response.Response	"Department not found"
//	@Router			/departments/{_id} [get]
func getDepartmentDoc() {}

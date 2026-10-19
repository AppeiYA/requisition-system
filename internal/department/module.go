package department

import (
	departmenthttp "github.com/AppeiYA/requisition-system/internal/department/adapters/in/http"
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
)

type DepartmentModule struct {
	departmentPortIn *departmentin.DepartmentPortIn
	departmentHandler *departmenthttp.DepartmentHandler
}

func NewDepartmentModule(
	departmentPortIn *departmentin.DepartmentPortIn,
	departmentHandler *departmenthttp.DepartmentHandler,
) *DepartmentModule {
	return &DepartmentModule{
		departmentPortIn: departmentPortIn,
		departmentHandler: departmentHandler,
	}
}

func (m *DepartmentModule) Handler() *departmenthttp.DepartmentHandler {
	return m.departmentHandler
}

func (m *DepartmentModule) PortIn() *departmentin.DepartmentPortIn {
	return m.departmentPortIn
}
package department

import (
	departmenthttp "github.com/AppeiYA/requisition-system/internal/department/adapters/in/http"
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	departmentout "github.com/AppeiYA/requisition-system/internal/department/ports/out"
)

type DepartmentModule struct {
	departmentPortIn  *departmentin.DepartmentPortIn
	departmentHandler *departmenthttp.DepartmentHandler
	membershipRepo    departmentout.MembershipRepository
}

func NewDepartmentModule(
	departmentPortIn *departmentin.DepartmentPortIn,
	departmentHandler *departmenthttp.DepartmentHandler,
	membershipRepo departmentout.MembershipRepository,
) *DepartmentModule {
	return &DepartmentModule{
		departmentPortIn:  departmentPortIn,
		departmentHandler: departmentHandler,
		membershipRepo:    membershipRepo,
	}
}

func (m *DepartmentModule) MembershipRepository() departmentout.MembershipRepository {
	return m.membershipRepo
}

func (m *DepartmentModule) Handler() *departmenthttp.DepartmentHandler {
	return m.departmentHandler
}

func (m *DepartmentModule) PortIn() *departmentin.DepartmentPortIn {
	return m.departmentPortIn
}

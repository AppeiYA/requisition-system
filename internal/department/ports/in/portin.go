package departmentin

type DepartmentPortIn struct {
	CreateMembershipUsecase  CreateMembershipUsecaseInt
	CreateDepartmentUsecase  CreateDepartmentUsecaseInt
	GetDepartmentByIDUsecase GetDepartmentByIDUsecaseInt
	GetDepartmentsUsecase    GetDepartmentsUsecaseInt
}

func NewDepartmentPortIn(
	createMembershipUsecase CreateMembershipUsecaseInt,
	createDepartmentUsecase CreateDepartmentUsecaseInt,
	getDepartmentByIDUsecase GetDepartmentByIDUsecaseInt,
	getDepartmentsUsecase GetDepartmentsUsecaseInt,
) *DepartmentPortIn {
	return &DepartmentPortIn{
		CreateMembershipUsecase:  createMembershipUsecase,
		CreateDepartmentUsecase:  createDepartmentUsecase,
		GetDepartmentByIDUsecase: getDepartmentByIDUsecase,
		GetDepartmentsUsecase:    getDepartmentsUsecase,
	}
}

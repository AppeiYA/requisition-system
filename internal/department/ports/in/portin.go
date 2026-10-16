package departmentin

type DepartmentPortIn struct {
	CreateMembershipUsecase CreateMembershipUsecaseInt
	CreateDepartmentUsecase CreateDepartmentUsecaseInt
	GetDepartmentByIDUsecase GetDepartmentByIDUsecaseInt
}

func NewDepartmentPortIn(
	createMembershipUsecase CreateMembershipUsecaseInt,
	createDepartmentUsecase CreateDepartmentUsecaseInt,
	getDepartmentByIDUsecase GetDepartmentByIDUsecaseInt,
) *DepartmentPortIn {
	return &DepartmentPortIn{
		CreateMembershipUsecase: createMembershipUsecase,
		CreateDepartmentUsecase: createDepartmentUsecase,
		GetDepartmentByIDUsecase: getDepartmentByIDUsecase,
	}
}
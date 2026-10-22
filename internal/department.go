package app

import (
	"github.com/AppeiYA/requisition-system/internal/department"
	departmenthttp "github.com/AppeiYA/requisition-system/internal/department/adapters/in/http"
	departmentpostgres "github.com/AppeiYA/requisition-system/internal/department/adapters/out/postgres"
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	departmentusecase "github.com/AppeiYA/requisition-system/internal/department/usecase"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/cache"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
)

func (a *App) buildDepartment(
	baseRepo db.Repository,
	idGen identityout.IDGenerator,
	cache cache.Cache,
) *department.DepartmentModule {
	departmentRepo := departmentpostgres.NewDepartmentRepository(baseRepo, cache)
	membershipRepo := departmentpostgres.NewMembershipRepository(baseRepo)

	createDepartmentUC := departmentusecase.NewCreateDepartmentUsecase(departmentRepo, idGen)
	getDepartmentByIDUC := departmentusecase.NewGetDepartmentByIDUsecase(departmentRepo)
	getDepartmentsUC := departmentusecase.NewGetDepartmentsUsecase(departmentRepo)
	createMembershipUC := departmentusecase.NewCreateMembershipUsecase(membershipRepo)

	departmentPortIn := departmentin.NewDepartmentPortIn(createMembershipUC, createDepartmentUC, getDepartmentByIDUC, getDepartmentsUC)

	handler := departmenthttp.NewDepartmentHandler(departmentPortIn)

	return department.NewDepartmentModule(
		departmentPortIn,
		handler,
	)
}

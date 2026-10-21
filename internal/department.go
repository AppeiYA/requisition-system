package app

import (
	"github.com/AppeiYA/requisition-system/internal/department"
	departmenthttp "github.com/AppeiYA/requisition-system/internal/department/adapters/in/http"
	departmentpostgres "github.com/AppeiYA/requisition-system/internal/department/adapters/out/postgres"
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	departmentusecase "github.com/AppeiYA/requisition-system/internal/department/usecase"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
)

func (a *App) buildDepartment(
	database *db.DB,
	idGen identityout.IDGenerator,
) *department.DepartmentModule {
	departmentRepo := departmentpostgres.NewDepartmentRepository(db.NewRepository(database))
	membershipRepo := departmentpostgres.NewMembershipRepository(db.NewRepository(database))

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

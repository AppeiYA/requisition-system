package testutil

import (
	"testing"

	departmenthttp "github.com/AppeiYA/requisition-system/internal/department/adapters/in/http"
	departmentpostgres "github.com/AppeiYA/requisition-system/internal/department/adapters/out/postgres"
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	departmentusecase "github.com/AppeiYA/requisition-system/internal/department/usecase"
	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/generator"
	redisinfra "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/gin-gonic/gin"
)

type DepartmentHTTP struct {
	*IdentityHTTP
	Router               *gin.Engine
	DepartmentRepository *departmentpostgres.DepartmentRepository
}

func (s *Suite) SetupDepartmentHTTP(t *testing.T) *DepartmentHTTP {
	t.Helper()

	identity := s.SetupIdentityHTTP(t)
	repository := departmentpostgres.NewDepartmentRepository(
		db.NewRepository(s.AppDB),
		redisinfra.NewCache(s.Redis.Client),
	)

	createDepartment := departmentusecase.NewCreateDepartmentUsecase(
		repository,
		&generator.IDGenerator{},
	)
	getDepartmentByID := departmentusecase.NewGetDepartmentByIDUsecase(repository)
	getDepartments := departmentusecase.NewGetDepartmentsUsecase(repository)
	ports := departmentin.NewDepartmentPortIn(
		nil,
		createDepartment,
		getDepartmentByID,
		getDepartments,
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	authMiddleware := identityhttp.NewAuthMiddleware(identity.AccessTokens, identity.UserRepository)
	departmenthttp.NewDepartmentHandler(ports).Register(
		router.Group("/api/v1"),
		authMiddleware.Handle(),
	)

	return &DepartmentHTTP{
		IdentityHTTP:         identity,
		Router:               router,
		DepartmentRepository: repository,
	}
}

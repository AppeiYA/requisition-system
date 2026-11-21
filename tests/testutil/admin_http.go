package testutil

import (
	"testing"

	adminhttp "github.com/AppeiYA/requisition-system/internal/admin/adapters/in/http"
	adminexternal "github.com/AppeiYA/requisition-system/internal/admin/adapters/out/external"
	adminmocks "github.com/AppeiYA/requisition-system/internal/admin/mocks"
	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	adminusecase "github.com/AppeiYA/requisition-system/internal/admin/usecase"
	departmentpostgres "github.com/AppeiYA/requisition-system/internal/department/adapters/out/postgres"
	departmentusecase "github.com/AppeiYA/requisition-system/internal/department/usecase"
	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	identityusecase "github.com/AppeiYA/requisition-system/internal/identity/usecase"
	invitationpostgres "github.com/AppeiYA/requisition-system/internal/invitation/adapters/out/postgres"
	invitationusecase "github.com/AppeiYA/requisition-system/internal/invitation/usecase"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/generator"
	redisinfra "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/gin-gonic/gin"
)

type AdminHTTP struct {
	*IdentityHTTP
	Router               *gin.Engine
	InvitationRepository *invitationpostgres.InvitationRepository
	DepartmentRepository *departmentpostgres.DepartmentRepository
	TaskQueue            *adminmocks.MockTaskQueue
}

func (s *Suite) SetupAdminHTTP(t *testing.T) *AdminHTTP {
	t.Helper()

	identity := s.SetupIdentityHTTP(t)
	baseRepository := db.NewRepository(s.AppDB)
	cache := redisinfra.NewCache(s.Redis.Client)

	departmentRepository := departmentpostgres.NewDepartmentRepository(baseRepository, cache)
	getDepartmentByID := departmentusecase.NewGetDepartmentByIDUsecase(departmentRepository)

	invitationRepository := invitationpostgres.NewInvitationRepository(baseRepository)
	saveInvitation := invitationusecase.NewSaveInvitationUsecase(
		invitationRepository,
		&generator.IDGenerator{},
	)

	getUserByEmail := identityusecase.NewGetUserByEmailUsecase(identity.UserRepository)
	adminIdentity := adminexternal.NewAdminIdentityAdapter(getUserByEmail)
	adminInvitation := adminexternal.NewAdminInvitationAdapter(saveInvitation)
	adminDepartment := adminexternal.NewAdminDepartmentAdapter(getDepartmentByID)
	taskQueue := adminmocks.NewMockTaskQueue()
	transactionManager := db.NewTransactionManager(s.AppDB)

	inviteMember := adminusecase.NewInviteMemberUsecase(
		taskQueue,
		adminIdentity,
		adminInvitation,
		adminDepartment,
		transactionManager,
		adminmocks.NewMockTelemetry(),
	)
	handler := adminhttp.NewHandler(adminin.NewAdminPortIn(inviteMember))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	authMiddleware := identityhttp.NewAuthMiddleware(identity.AccessTokens, identity.UserRepository)
	handler.Register(router.Group("/api/v1"), authMiddleware.IsAdmin())

	return &AdminHTTP{
		IdentityHTTP:         identity,
		Router:               router,
		InvitationRepository: invitationRepository,
		DepartmentRepository: departmentRepository,
		TaskQueue:            taskQueue,
	}
}

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
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityusecase "github.com/AppeiYA/requisition-system/internal/identity/usecase"
	invitationpostgres "github.com/AppeiYA/requisition-system/internal/invitation/adapters/out/postgres"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationusecase "github.com/AppeiYA/requisition-system/internal/invitation/usecase"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/audit"
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
	// IdentityAdminUsers lets tests exercise domain guards that cannot be
	// reached through HTTP: an authenticated caller is always an active admin,
	// so the last-admin guard needs a direct use case call.
	IdentityAdminUsers *identityin.AdminUsersPortIn
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

	invitationAuditRecorder := audit.NewPostgresRecorder(baseRepository, &generator.IDGenerator{})
	adminInvitations := &invitationin.AdminInvitationsPortIn{
		ListInvitations:  invitationusecase.NewAdminListInvitationsUsecase(invitationRepository),
		GetInvitation:    invitationusecase.NewAdminGetInvitationUsecase(invitationRepository),
		ResendInvitation: invitationusecase.NewAdminResendInvitationUsecase(invitationRepository, transactionManager, invitationAuditRecorder),
		RevokeInvitation: invitationusecase.NewAdminRevokeInvitationUsecase(invitationRepository, transactionManager, invitationAuditRecorder),
	}
	adminInvitationManagement := adminexternal.NewAdminInvitationManagementAdapter(adminInvitations)

	telemetry := adminmocks.NewMockTelemetry()
	inviteMember := adminusecase.NewInviteMemberUsecase(
		taskQueue,
		adminIdentity,
		adminInvitation,
		adminDepartment,
		transactionManager,
		telemetry,
	)

	auditRecorder := audit.NewPostgresRecorder(baseRepository, &generator.IDGenerator{})
	identityAdminUsers := &identityin.AdminUsersPortIn{
		ListUsers:        identityusecase.NewListUsersUsecase(identity.UserRepository),
		GetUser:          identityusecase.NewGetUserUsecase(identity.UserRepository),
		UpdateUserStatus: identityusecase.NewUpdateUserStatusUsecase(identity.UserRepository, identity.RefreshSessionRepository, transactionManager, auditRecorder),
		UpdateUserRole:   identityusecase.NewUpdateUserRoleUsecase(identity.UserRepository, identity.RefreshSessionRepository, transactionManager, auditRecorder),
		RevokeUserSession: identityusecase.NewRevokeUserSessionsUsecase(
			identity.UserRepository,
			identity.RefreshSessionRepository,
			auditRecorder,
		),
	}
	userAdminPort := adminexternal.NewAdminIdentityUserAdapter(identityAdminUsers)

	handler := adminhttp.NewHandler(adminin.NewAdminPortInWithAll(
		inviteMember,
		adminusecase.NewListUsersUsecase(userAdminPort),
		adminusecase.NewGetUserUsecase(userAdminPort),
		adminusecase.NewUpdateUserStatusUsecase(userAdminPort, telemetry),
		adminusecase.NewUpdateUserRoleUsecase(userAdminPort, telemetry),
		adminusecase.NewRevokeUserSessionsUsecase(userAdminPort, telemetry),
		adminusecase.NewListInvitationsUsecase(adminInvitationManagement),
		adminusecase.NewGetInvitationUsecase(adminInvitationManagement),
		adminusecase.NewResendInvitationUsecase(adminInvitationManagement, adminDepartment, taskQueue, telemetry),
		adminusecase.NewRevokeInvitationUsecase(adminInvitationManagement, telemetry),
	))

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
		IdentityAdminUsers:   identityAdminUsers,
	}
}

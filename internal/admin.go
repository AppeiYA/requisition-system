package app

import (
	"github.com/AppeiYA/requisition-system/internal/admin"
	adminhttp "github.com/AppeiYA/requisition-system/internal/admin/adapters/in/http"
	adminworkers "github.com/AppeiYA/requisition-system/internal/admin/adapters/in/worker"
	adminemail "github.com/AppeiYA/requisition-system/internal/admin/adapters/out/email"
	adminexternal "github.com/AppeiYA/requisition-system/internal/admin/adapters/out/external"
	adminworker "github.com/AppeiYA/requisition-system/internal/admin/adapters/out/worker"
	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	adminusecase "github.com/AppeiYA/requisition-system/internal/admin/usecase"
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/email"
	"github.com/hibiken/asynq"
)

func (a *App) buildAdminHandler(
	cfg *config.Config,
	txManager *db.TransactionManager,
	getUserByEmailUC identityin.GetUserByEmailUsecaseInt,
	adminUsersUC *identityin.AdminUsersPortIn,
	saveInvitationUC invitationin.SaveInvitationUsecaseInt,
	adminInvitationsUC *invitationin.AdminInvitationsPortIn,
	getDepartmentByIDUC departmentin.GetDepartmentByIDUsecaseInt,
	asynqClient *asynq.Client,
	telemetry adminout.Telemetry,
) *admin.AdminModule {
	taskQueue := adminworker.NewTaskQueue(asynqClient, cfg.Worker)
	identityPort := adminexternal.NewAdminIdentityAdapter(getUserByEmailUC)
	adminInvitationPort := adminexternal.NewAdminInvitationAdapter(saveInvitationUC)
	adminInvitationManagementPort := adminexternal.NewAdminInvitationManagementAdapter(adminInvitationsUC)
	adminDepartmentPort := adminexternal.NewAdminDepartmentAdapter(getDepartmentByIDUC)

	inviteUC := adminusecase.NewInviteMemberUsecase(taskQueue, identityPort, adminInvitationPort, adminDepartmentPort, txManager, telemetry)

	userAdminPort := adminexternal.NewAdminIdentityUserAdapter(adminUsersUC)
	listUsersUC := adminusecase.NewListUsersUsecase(userAdminPort)
	getUserUC := adminusecase.NewGetUserUsecase(userAdminPort)
	updateStatusUC := adminusecase.NewUpdateUserStatusUsecase(userAdminPort, telemetry)
	updateRoleUC := adminusecase.NewUpdateUserRoleUsecase(userAdminPort, telemetry)
	revokeSessionsUC := adminusecase.NewRevokeUserSessionsUsecase(userAdminPort, telemetry)

	listInvitationsUC := adminusecase.NewListInvitationsUsecase(adminInvitationManagementPort)
	getInvitationUC := adminusecase.NewGetInvitationUsecase(adminInvitationManagementPort)
	resendInvitationUC := adminusecase.NewResendInvitationUsecase(adminInvitationManagementPort, adminDepartmentPort, taskQueue, telemetry)
	revokeInvitationUC := adminusecase.NewRevokeInvitationUsecase(adminInvitationManagementPort, telemetry)

	adminPortIn := adminin.NewAdminPortInWithAll(
		inviteUC,
		listUsersUC,
		getUserUC,
		updateStatusUC,
		updateRoleUC,
		revokeSessionsUC,
		listInvitationsUC,
		getInvitationUC,
		resendInvitationUC,
		revokeInvitationUC,
	)
	adminHandler := adminhttp.NewHandler(adminPortIn)

	return admin.NewAdminModule(adminPortIn, adminHandler)
}

func (w *Worker) buildAdminWorker(
	cfg *config.Config,
	invitationRepo invitationout.InvitationRepository,
	sender *email.Sender,
	telemetry adminout.Telemetry,
) *admin.AdminWorkerModule {
	emailPort := adminemail.NewInvitationSender(sender, cfg.Frontend)

	sendInvitationEmailUsecase := adminusecase.NewSendInvitationEmailUsecase(emailPort, telemetry)

	sendMailHandler := adminworkers.NewSendMailHandler(sendInvitationEmailUsecase, invitationRepo)

	return admin.NewAdminWorkerModule(sendMailHandler)
}

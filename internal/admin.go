package app

import (
	"github.com/AppeiYA/requisition-system/internal/admin"
	adminhttp "github.com/AppeiYA/requisition-system/internal/admin/adapters/in/http"
	adminworkers "github.com/AppeiYA/requisition-system/internal/admin/adapters/in/worker"
	adminemail "github.com/AppeiYA/requisition-system/internal/admin/adapters/out/email"
	adminexternal "github.com/AppeiYA/requisition-system/internal/admin/adapters/out/external"
	adminworker "github.com/AppeiYA/requisition-system/internal/admin/adapters/out/worker"
	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
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
	saveInvitationUC invitationin.SaveInvitationUsecaseInt,
	getDepartmentByIDUC departmentin.GetDepartmentByIDUsecaseInt,
	asynqClient *asynq.Client,
) *admin.AdminModule {
	taskQueue := adminworker.NewTaskQueue(asynqClient, cfg.Worker)
	identityPort := adminexternal.NewAdminIdentityAdapter(getUserByEmailUC)
	adminInvitationPort := adminexternal.NewAdminInvitationAdapter(saveInvitationUC)
	adminDepartmentPort := adminexternal.NewAdminDepartmentAdapter(getDepartmentByIDUC)

	inviteUC := adminusecase.NewInviteMemberUsecase(taskQueue, identityPort, adminInvitationPort, adminDepartmentPort, txManager)

	adminPortIn := adminin.NewAdminPortIn(inviteUC)
	adminHandler := adminhttp.NewHandler(adminPortIn)

	return admin.NewAdminModule(adminPortIn, adminHandler)
}

func (w *Worker) buildAdminWorker(cfg *config.Config, invitationRepo invitationout.InvitationRepository) *admin.AdminWorkerModule {
	emailPort := adminemail.NewInvitationSender(email.NewSender(cfg.Email), cfg.Frontend)

	sendInvitationEmailUsecase := adminusecase.NewSendInvitationEmailUsecase(emailPort)

	sendMailHandler := adminworkers.NewSendMailHandler(sendInvitationEmailUsecase, invitationRepo)

	return admin.NewAdminWorkerModule(sendMailHandler)
}

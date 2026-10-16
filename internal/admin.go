package app

import (
	"github.com/AppeiYA/requisition-system/internal/admin"
	adminhttp "github.com/AppeiYA/requisition-system/internal/admin/adapters/in/http"
	adminemail "github.com/AppeiYA/requisition-system/internal/admin/adapters/out/email"
	adminexternal "github.com/AppeiYA/requisition-system/internal/admin/adapters/out/external"
	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	adminusecase "github.com/AppeiYA/requisition-system/internal/admin/usecase"
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/email"
)

func (a *App) buildAdminHandler(
	cfg *config.Config,
	txManager *db.TransactionManager,
	getUserByEmailUC identityin.GetUserByEmailUsecaseInt,
	saveInvitationUC invitationin.SaveInvitationUsecaseInt,
	getInvitationUsecase invitationin.GetInvitationUsecaseInt,
	getDepartmentByIDUC departmentin.GetDepartmentByIDUsecaseInt,
	emailSender *email.Sender,
) *admin.AdminModule {

	emailPort := adminemail.NewInvitationSender(emailSender, cfg.Frontend)
	identityPort := adminexternal.NewAdminIdentityAdapter(getUserByEmailUC)
	adminInvitationPort := adminexternal.NewAdminInvitationAdapter(saveInvitationUC, getInvitationUsecase)
	adminDepartmentPort := adminexternal.NewAdminDepartmentAdapter(getDepartmentByIDUC)


	inviteUC := adminusecase.NewInviteMemberUsecase(emailPort, identityPort, adminInvitationPort, adminDepartmentPort, txManager)
	
	adminPortIn := adminin.NewAdminPortIn(inviteUC)
	adminHandler := adminhttp.NewHandler(adminPortIn)
	

	return admin.NewAdminModule(adminPortIn, adminHandler)
}

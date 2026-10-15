package app

import (
	adminhttp "github.com/AppeiYA/requisition-system/internal/admin/adapters/in/http"
	adminemail "github.com/AppeiYA/requisition-system/internal/admin/adapters/out/email"
	adminexternal "github.com/AppeiYA/requisition-system/internal/admin/adapters/out/external"
	adminusecase "github.com/AppeiYA/requisition-system/internal/admin/usecase"
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/email"
)

func (a *App) buildAdminHandler(
	cfg *config.Config,
	txManager *db.TransactionManager,
	getUserByEmailUC identityin.GetUserByEmailUsecaseInt,
	saveInvitationUC invitationin.SaveInvitationUsecaseInt,
	invRepo invitationout.InvitationRepository,
	getDepartmentByIDUC departmentin.GetDepartmentByIDUsecaseInt,
	emailSender *email.Sender,
) *adminhttp.Handler {
	emailPort := adminemail.NewInvitationSender(emailSender, cfg.Frontend)
	identityPort := adminexternal.NewAdminIdentityAdapter(getUserByEmailUC)
	invitationPort := adminexternal.NewAdminInvitationAdapter(saveInvitationUC, invRepo)
	departmentPort := adminexternal.NewAdminDepartmentAdapter(getDepartmentByIDUC)

	inviteUC := adminusecase.NewInviteMemberUsecase(emailPort, identityPort, invitationPort, departmentPort, txManager)

	return adminhttp.NewHandler(inviteUC)
}

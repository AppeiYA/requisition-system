package app

import (
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/invitation"
	invitationhttp "github.com/AppeiYA/requisition-system/internal/invitation/adapters/in/http"
	invitationexternal "github.com/AppeiYA/requisition-system/internal/invitation/adapters/out/external"
	invitationpostgres "github.com/AppeiYA/requisition-system/internal/invitation/adapters/out/postgres"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationusecase "github.com/AppeiYA/requisition-system/internal/invitation/usecase"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
)

func (a *App) buildInvitation(
	database *db.DB,
	txManager *db.TransactionManager,
	registerUserUC identityin.RegisterUserUseCaseInt,
	createMembershipUC departmentin.CreateMembershipUsecaseInt,
	idGen identityout.IDGenerator,
) (
	*invitation.InvitationModule,
) {
	invitationRepo := invitationpostgres.NewInvitationRepository(db.NewRepository(database))

	// The identity port lets the invitation module register the invited user
	// without depending on the identity module's concrete use cases.
	invitationIdentityPort := invitationexternal.NewInviteIdentityAdapter(registerUserUC)

	// The department port lets the invitation module create the membership on
	// acceptance without depending on the department module's domain.
	invitationDepartmentPort := invitationexternal.NewInviteDepartmentAdapter(createMembershipUC)

	acceptInvitationUsecase := invitationusecase.NewAcceptInvitationUsecase(
		invitationRepo,
		invitationIdentityPort,
		invitationDepartmentPort,
		txManager,
	)
	validateInvitationUsecase := invitationusecase.NewValidateInvitationUsecase(invitationRepo)
	saveInvitationUsecase := invitationusecase.NewSaveInvitationUsecase(invitationRepo, idGen)
	getInvitationUsecase := invitationusecase.NewGetInvitationUsecase(invitationRepo)

	invitationPortIn := invitationin.NewInvitationPortIn(saveInvitationUsecase, validateInvitationUsecase, acceptInvitationUsecase, getInvitationUsecase)

	handler := invitationhttp.NewInvitationHandler(invitationPortIn)

	return invitation.NewInvitationModule(
		invitationPortIn,
		handler,
	)
}

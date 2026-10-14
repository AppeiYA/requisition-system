package app

import (
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	invitationhttp "github.com/AppeiYA/requisition-system/internal/invitation/adapters/in/http"
	invitationexternal "github.com/AppeiYA/requisition-system/internal/invitation/adapters/out/external"
	invitationpostgres "github.com/AppeiYA/requisition-system/internal/invitation/adapters/out/postgres"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	invitationusecase "github.com/AppeiYA/requisition-system/internal/invitation/usecase"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
)

func (a *App) buildInvitation(
	database *db.DB,
	txManager *db.TransactionManager,
	registerUserUC identityin.RegisterUserUseCaseInt,
	idGen identityout.IDGenerator,
) (
	*invitationhttp.InvitationHandler,
	invitationout.InvitationRepository,
	invitationin.SaveInvitationUsecaseInt,
) {
	invitationRepo := invitationpostgres.NewInvitationRepository(db.NewRepository(database))

	saveInvitationUsecase := invitationusecase.NewSaveInvitationUsecase(invitationRepo, idGen)

	// The identity port lets the invitation module register the invited user
	// without depending on the identity module's concrete use cases.
	identityPort := invitationexternal.NewInviteIdentityAdapter(registerUserUC)

	acceptInvitationUsecase := invitationusecase.NewAcceptInvitationUsecase(
		invitationRepo,
		identityPort,
		txManager,
	)
	validateInvitationUsecase := invitationusecase.NewValidateInvitationUsecase(invitationRepo)

	handler := invitationhttp.NewInvitationHandler(acceptInvitationUsecase, validateInvitationUsecase)

	return handler, invitationRepo, saveInvitationUsecase
}

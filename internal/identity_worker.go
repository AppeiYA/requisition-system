package app

import (
	"github.com/AppeiYA/requisition-system/internal/identity"
	identityworkerin "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/worker"
	identityemail "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/email"
	identitypostgres "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	identityusecase "github.com/AppeiYA/requisition-system/internal/identity/usecase"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/email"
)

// buildIdentityWorker wires the identity tasks to their adapters. It is the
// worker-process counterpart of buildIdentity.
func (w *Worker) buildIdentityWorker(
	cfg *config.Config,
	database *db.DB,
	sender *email.Sender,
	telemetry identityout.Telemetry,
) *identity.IdentityWorkerModule {
	tokenRepo := identitypostgres.NewPasswordResetTokenRepository(db.NewRepository(database))
	emailPort := identityemail.NewPasswordResetSender(sender, cfg.Frontend, cfg.PasswordReset)

	sendEmailUsecase := identityusecase.NewSendPasswordResetEmailUsecase(emailPort, telemetry)
	cleanupUsecase := identityusecase.NewCleanupPasswordResetTokensUsecase(tokenRepo, cfg.PasswordReset)

	sendHandler := identityworkerin.NewSendPasswordResetEmailHandler(sendEmailUsecase)
	cleanupHandler := identityworkerin.NewCleanupPasswordResetTokensHandler(cleanupUsecase)

	return identity.NewIdentityWorkerModule(sendHandler, cleanupHandler)
}

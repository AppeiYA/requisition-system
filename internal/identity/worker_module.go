package identity

import (
	identityworkerin "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/worker"
)

// IdentityWorkerModule holds the inbound adapters of the identity tasks the
// worker process serves. It mirrors AdminWorkerModule.
type IdentityWorkerModule struct {
	sendPasswordResetEmailHandler     *identityworkerin.SendPasswordResetEmailHandler
	cleanupPasswordResetTokensHandler *identityworkerin.CleanupPasswordResetTokensHandler
}

func NewIdentityWorkerModule(
	sendPasswordResetEmailHandler *identityworkerin.SendPasswordResetEmailHandler,
	cleanupPasswordResetTokensHandler *identityworkerin.CleanupPasswordResetTokensHandler,
) *IdentityWorkerModule {
	return &IdentityWorkerModule{
		sendPasswordResetEmailHandler:     sendPasswordResetEmailHandler,
		cleanupPasswordResetTokensHandler: cleanupPasswordResetTokensHandler,
	}
}

func (m *IdentityWorkerModule) SendPasswordResetEmailHandler() *identityworkerin.SendPasswordResetEmailHandler {
	return m.sendPasswordResetEmailHandler
}

func (m *IdentityWorkerModule) CleanupPasswordResetTokensHandler() *identityworkerin.CleanupPasswordResetTokensHandler {
	return m.cleanupPasswordResetTokensHandler
}

package admin

import (
	adminworkers "github.com/AppeiYA/requisition-system/internal/admin/adapters/in/worker"
)

// AdminWorkerModule is the worker-process counterpart of AdminModule: it
// carries the adapters that run inside the asynq worker instead of the
// HTTP ones. Each domain exposes one so the composition root can register
// its tasks without knowing the domain's internals.
type AdminWorkerModule struct {
	sendMailHandler *adminworkers.SendMailHandler
}

func NewAdminWorkerModule(sendMailHandler *adminworkers.SendMailHandler) *AdminWorkerModule {
	return &AdminWorkerModule{
		sendMailHandler: sendMailHandler,
	}
}

// SendMailHandler returns the handler for the invitation email task; the
// composition root registers it on the asynq mux.
func (m *AdminWorkerModule) SendMailHandler() *adminworkers.SendMailHandler {
	return m.sendMailHandler
}

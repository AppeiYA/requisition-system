package attachment

import (
	attachmentworker "github.com/AppeiYA/requisition-system/internal/attachment/adapters/in/worker"
)

// AttachmentWorkerModule exposes only the handlers the worker process needs.
type AttachmentWorkerModule struct {
	reconcileHandler *attachmentworker.ReconcileOutboxHandler
	cleanupHandler   *attachmentworker.CleanupAttachmentsHandler
}

func NewAttachmentWorkerModule(
	reconcileHandler *attachmentworker.ReconcileOutboxHandler,
	cleanupHandler *attachmentworker.CleanupAttachmentsHandler,
) *AttachmentWorkerModule {
	return &AttachmentWorkerModule{
		reconcileHandler: reconcileHandler,
		cleanupHandler:   cleanupHandler,
	}
}

func (m *AttachmentWorkerModule) ReconcileOutboxHandler() *attachmentworker.ReconcileOutboxHandler {
	return m.reconcileHandler
}

func (m *AttachmentWorkerModule) CleanupHandler() *attachmentworker.CleanupAttachmentsHandler {
	return m.cleanupHandler
}

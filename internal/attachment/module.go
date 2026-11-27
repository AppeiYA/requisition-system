package attachment

import (
	attachmenthttp "github.com/AppeiYA/requisition-system/internal/attachment/adapters/in/http"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
)

type AttachmentModule struct {
	portIn  *attachmentin.AttachmentPortIn
	handler *attachmenthttp.AttachmentHandler
}

func NewAttachmentModule(
	portIn *attachmentin.AttachmentPortIn,
	handler *attachmenthttp.AttachmentHandler,
) *AttachmentModule {
	return &AttachmentModule{portIn: portIn, handler: handler}
}

func (m *AttachmentModule) PortIn() *attachmentin.AttachmentPortIn { return m.portIn }

func (m *AttachmentModule) Handler() *attachmenthttp.AttachmentHandler { return m.handler }

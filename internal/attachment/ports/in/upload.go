package attachmentin

import (
	"context"
	"io"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
)

type UploadAttachmentCommand struct {
	ActorID   string
	ActorRole string
	RefName   string
	RefID     string
	FileName  string
	Body      io.Reader
	Size      int64
}

type UploadAttachmentUsecaseInt interface {
	Execute(ctx context.Context, cmd UploadAttachmentCommand) (*domain.Attachment, error)
}

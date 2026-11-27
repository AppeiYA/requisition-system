package attachmentin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
)

type GetAttachmentQuery struct {
	AttachmentID string
	ActorID      string
	ActorRole    string
}

type GetAttachmentUsecaseInt interface {
	Execute(ctx context.Context, query GetAttachmentQuery) (*domain.Attachment, error)
}

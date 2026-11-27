package attachmentin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
)

type ListAttachmentsQuery struct {
	ActorID   string
	ActorRole string
	RefName   string
	RefID     string
	Limit     int
	Offset    int
}

type ListAttachmentsUsecaseInt interface {
	Execute(ctx context.Context, query ListAttachmentsQuery) ([]*domain.Attachment, error)
}

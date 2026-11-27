package attachmentin

import (
	"context"
	"io"
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
)

type DownloadAttachmentQuery struct {
	AttachmentID string
	ActorID      string
	ActorRole    string
}

// Download is either a stream of the stored bytes or a time-limited Location
// when the storage adapter can sign URLs. It is never both.
type Download struct {
	Attachment *domain.Attachment
	Stream     io.ReadSeekCloser
	Location   string
	ExpiresAt  time.Time
}

type DownloadAttachmentUsecaseInt interface {
	Execute(ctx context.Context, query DownloadAttachmentQuery) (*Download, error)
}

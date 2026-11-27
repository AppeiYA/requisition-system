package attachmentin

import (
	"context"
	"io"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
)

// StageUploadCommand uploads a file before a requisition exists. Only the
// authenticated actor is required; parent authorisation happens later, when
// the requisition is created.
type StageUploadCommand struct {
	ActorID  string
	FileName string
	Body     io.Reader
	Size     int64
}

type StageUploadUsecaseInt interface {
	Execute(ctx context.Context, cmd StageUploadCommand) (*domain.StagedUpload, error)
}

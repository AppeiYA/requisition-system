package attachmentin

import "context"

type CleanupAttachmentsUsecaseInt interface {
	Execute(ctx context.Context) error
}

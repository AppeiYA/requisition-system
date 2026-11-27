package attachmentin

import "context"

type DeleteAttachmentCommand struct {
	AttachmentID string
	ActorID      string
	ActorRole    string
}

type DeleteAttachmentUsecaseInt interface {
	Execute(ctx context.Context, cmd DeleteAttachmentCommand) error
}

// DeleteForRefCommand is used internally when a parent is removed. It has no
// actor because the caller has already authorised that operation.
type DeleteForRefCommand struct {
	RefName string
	RefID   string
}

type DeleteAttachmentsForRefUsecaseInt interface {
	Execute(ctx context.Context, cmd DeleteForRefCommand) (int, error)
}

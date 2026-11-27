package attachmentin

import "context"

// ClaimStagedAttachmentTarget binds an already uploaded file to the real
// parent ID allocated by requisition creation. RefName is the attachment ref
// name, never a catalogue item ID.
type ClaimStagedAttachmentTarget struct {
	StagedUploadID string
	ActorID        string
	RefName        string
	RefID          string
}

type ClaimStagedAttachmentsUsecaseInt interface {
	Claim(ctx context.Context, targets []ClaimStagedAttachmentTarget) error
}

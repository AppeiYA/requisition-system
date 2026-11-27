package requisitionout

import "context"

// StagedAttachmentTarget names the real parent of one claimed upload. RefName
// is either "requisition" or "requisition_item"; RefID is the ID allocated
// inside the creation transaction, never a client-supplied catalogue ID.
type StagedAttachmentTarget struct {
	UploadID string
	ActorID  string
	RefName  string
	RefID    string
}

// StagedAttachmentRegistrar is implemented by the attachment module. It is
// called after the requisition and its line IDs exist but inside the same
// transaction, so a bad token rolls the whole creation back.
type StagedAttachmentRegistrar interface {
	RegisterStagedAttachments(ctx context.Context, targets []StagedAttachmentTarget) error
}

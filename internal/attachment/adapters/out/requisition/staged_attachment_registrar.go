// Package requisitionadapter implements requisition's outbound port for
// creation-time attachments. It translates the requisition's neutral target
// list into attachment use-case inputs; no requisition domain type crosses
// into the attachment module.
package requisitionadapter

import (
	"context"

	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

// StagedAttachmentRegistrar is the adapter used by requisition creation. The
// usecase call inherits the requisition's transaction context, so every claim,
// metadata row and outbox row commits or rolls back together.
type StagedAttachmentRegistrar struct {
	claim attachmentin.ClaimStagedAttachmentsUsecaseInt
}

func NewStagedAttachmentRegistrar(
	claim attachmentin.ClaimStagedAttachmentsUsecaseInt,
) *StagedAttachmentRegistrar {
	return &StagedAttachmentRegistrar{claim: claim}
}

func (r *StagedAttachmentRegistrar) RegisterStagedAttachments(
	ctx context.Context,
	targets []requisitionout.StagedAttachmentTarget,
) error {
	converted := make([]attachmentin.ClaimStagedAttachmentTarget, 0, len(targets))
	for _, target := range targets {
		converted = append(converted, attachmentin.ClaimStagedAttachmentTarget{
			StagedUploadID: target.UploadID,
			ActorID:        target.ActorID,
			RefName:        target.RefName,
			RefID:          target.RefID,
		})
	}
	return r.claim.Claim(ctx, converted)
}

var _ requisitionout.StagedAttachmentRegistrar = (*StagedAttachmentRegistrar)(nil)

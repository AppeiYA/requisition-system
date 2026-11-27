// Package external implements attachment out ports by calling the modules
// that own the referenced resources. It keeps the attachment module ignorant
// of requisition domain types.
package external

import (
	"context"

	attachmentdomain "github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
)

type RequisitionAccessChecker struct {
	requisitionPortIn *requisitionin.RequisitionPortIn
}

func NewRequisitionAccessChecker(requisitionPortIn *requisitionin.RequisitionPortIn) *RequisitionAccessChecker {
	return &RequisitionAccessChecker{requisitionPortIn: requisitionPortIn}
}

func (c *RequisitionAccessChecker) CanAccess(
	ctx context.Context,
	refName, refID, actorID, actorRole string,
	_ attachmentout.Action,
) error {
	switch attachmentdomain.AttachmentRefName(refName) {
	case attachmentdomain.RefNameRequisition:
		_, err := c.requisitionPortIn.GetRequisitionUsecase.Execute(ctx, requisitionin.GetRequisitionQuery{
			RequisitionID: refID,
			ActorID:       actorID,
			ActorRole:     actorRole,
		})
		return err
	case attachmentdomain.RefNameRequisitionItem:
		_, err := c.requisitionPortIn.GetRequisitionItemUsecase.Execute(ctx, requisitionin.GetRequisitionItemQuery{
			RequisitionItemID: refID,
			ActorID:           actorID,
			ActorRole:         actorRole,
		})
		return err
	default:
		return attachmentdomain.ErrInvalidAttachmentRefName
	}
}

var _ attachmentout.RefAccessChecker = (*RequisitionAccessChecker)(nil)

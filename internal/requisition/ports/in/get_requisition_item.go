package requisitionin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type GetRequisitionItemQuery struct {
	RequisitionItemID string
	ActorID           string
	ActorRole         string
}

// GetRequisitionItemUsecaseInt resolves a requisition line to its parent and
// applies the same access rule as reading the requisition itself.
type GetRequisitionItemUsecaseInt interface {
	Execute(ctx context.Context, query GetRequisitionItemQuery) (*domain.Requisition, error)
}

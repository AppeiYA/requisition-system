package requisitionin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type GetRequisitionQuery struct {
	RequisitionID string
	ActorID       string
	ActorRole     string
}

type GetRequisitionUsecaseInt interface {
	Execute(ctx context.Context, query GetRequisitionQuery) (*domain.Requisition, error)
}

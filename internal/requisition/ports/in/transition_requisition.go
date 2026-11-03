package requisitionin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type TransitionRequisitionCommand struct {
	RequisitionID string
	ActorID       string
	ActorRole     string
	ToStatus      string
	Reason        string
}

type TransitionRequisitionUsecaseInt interface {
	Execute(ctx context.Context, cmd TransitionRequisitionCommand) (*domain.Requisition, error)
}

package requisitionout

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

// RequisitionTransitionRepository is the persistence path dedicated to status
// transitions. Loading the requisition with a row lock and writing the header
// status together with its history entry keeps concurrent transitions
// serialized and atomic.
type RequisitionTransitionRepository interface {
	GetRequisitionByIDForUpdate(ctx context.Context, id string) (*domain.Requisition, error)
	UpdateRequisitionStatus(ctx context.Context, requisition *domain.Requisition, history *domain.RequisitionStatusHistory) error
}

package requisitionout

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type RequisitionEditRepository interface {
	GetRequisitionByIDForUpdate(context.Context, string) (*domain.Requisition, error)
	UpdateRequisitionAggregate(context.Context, *domain.Requisition) error
}

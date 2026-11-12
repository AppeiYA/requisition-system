package requisitionout

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

// RequisitionEditRepository is the write side used by the update use case.
//
// UpdateRequisitionAggregate must run in the same transaction as
// GetRequisitionByIDForUpdate: the lookup takes the requisition header row
// lock, and the aggregate update relies on that lock plus the transaction
// context to keep its persisted-line snapshot consistent.
type RequisitionEditRepository interface {
	GetRequisitionByIDForUpdate(context.Context, string) (*domain.Requisition, error)
	UpdateRequisitionAggregate(context.Context, *domain.Requisition) error
}

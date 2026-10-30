package requisitionout

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type RequisitionRepository interface {
	SaveRequisition(ctx context.Context, requisition *domain.Requisition) error
	GetRequisitionByID(ctx context.Context, id string) (*domain.Requisition, error)
	UpdateRequisition(ctx context.Context, requisition *domain.Requisition) error
	UpdateRequisitionAggregate(ctx context.Context, requisition *domain.Requisition) error
	ListRequisitionsByUserID(ctx context.Context, userID string, limit int, offset int) ([]*domain.Requisition, error)
	ListRequisitions(ctx context.Context, limit int, offset int) ([]*domain.Requisition, error)
}

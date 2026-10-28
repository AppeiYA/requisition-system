package requisitionin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type ListRequisitionsByUserQuery struct {
	UserID string
	Limit  int
	Offset int
}

type ListRequisitionsByUserUsecaseInt interface {
	Execute(ctx context.Context, query ListRequisitionsByUserQuery) ([]*domain.Requisition, error)
}

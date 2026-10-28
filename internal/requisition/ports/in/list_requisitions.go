package requisitionin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type ListRequisitionsQuery struct {
	Limit  int
	Offset int
}

type ListRequisitionsUsecaseInt interface {
	Execute(ctx context.Context, query ListRequisitionsQuery) ([]*domain.Requisition, error)
}

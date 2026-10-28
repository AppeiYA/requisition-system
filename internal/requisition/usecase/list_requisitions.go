package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

type ListRequisitionsUsecase struct {
	requisitionRepo requisitionout.RequisitionRepository
}

func NewListRequisitionsUsecase(requisitionRepo requisitionout.RequisitionRepository) *ListRequisitionsUsecase {
	return &ListRequisitionsUsecase{requisitionRepo: requisitionRepo}
}

func (uc *ListRequisitionsUsecase) Execute(ctx context.Context, query requisitionin.ListRequisitionsQuery) ([]*domain.Requisition, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	limit, offset := shareddomain.Clamp(query.Limit, query.Offset)
	return uc.requisitionRepo.ListRequisitions(ctx, limit, offset)
}

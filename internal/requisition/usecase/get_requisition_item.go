package usecase

import (
	"context"
	"strings"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

type GetRequisitionItemUsecase struct {
	requisitionRepo requisitionout.RequisitionRepository
}

func NewGetRequisitionItemUsecase(requisitionRepo requisitionout.RequisitionRepository) *GetRequisitionItemUsecase {
	return &GetRequisitionItemUsecase{requisitionRepo: requisitionRepo}
}

func (uc *GetRequisitionItemUsecase) Execute(ctx context.Context, query requisitionin.GetRequisitionItemQuery) (*domain.Requisition, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !strings.HasPrefix(query.RequisitionItemID, domain.RequisitionItemPrefixID+"-") {
		return nil, domain.ErrInvalidRequisitionItemID
	}

	requisitionID, err := uc.requisitionRepo.GetRequisitionIDByItemID(ctx, query.RequisitionItemID)
	if err != nil {
		return nil, err
	}
	requisition, err := uc.requisitionRepo.GetRequisitionByID(ctx, requisitionID)
	if err != nil {
		return nil, err
	}
	if !isAdminIdentityRole(query.ActorRole) && requisition.RequestedBy() != query.ActorID {
		return nil, ErrRequisitionAccessDenied
	}
	return requisition, nil
}

package usecase

import (
	"context"
	"strings"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

type GetRequisitionUsecase struct {
	requisitionRepo requisitionout.RequisitionRepository
}

func NewGetRequisitionUsecase(requisitionRepo requisitionout.RequisitionRepository) *GetRequisitionUsecase {
	return &GetRequisitionUsecase{requisitionRepo: requisitionRepo}
}

func (uc *GetRequisitionUsecase) Execute(ctx context.Context, query requisitionin.GetRequisitionQuery) (*domain.Requisition, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !strings.HasPrefix(query.RequisitionID, domain.RequisitionPrefixID) {
		return nil, domain.ErrInvalidRequisitionID
	}

	requisition, err := uc.requisitionRepo.GetRequisitionByID(ctx, query.RequisitionID)
	if err != nil {
		return nil, err
	}
	if !isAdminIdentityRole(query.ActorRole) && requisition.RequestedBy() != query.ActorID {
		return nil, ErrRequisitionAccessDenied
	}
	return requisition, nil
}

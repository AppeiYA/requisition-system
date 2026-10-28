package usecase

import (
	"context"
	"strings"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var ErrInvalidUserID = customerrors.NewBadRequestError("invalid user ID")

type ListRequisitionsByUserUsecase struct {
	requisitionRepo requisitionout.RequisitionRepository
	userLookup      requisitionout.UserLookup
}

func NewListRequisitionsByUserUsecase(requisitionRepo requisitionout.RequisitionRepository, userLookup requisitionout.UserLookup) *ListRequisitionsByUserUsecase {
	return &ListRequisitionsByUserUsecase{requisitionRepo: requisitionRepo, userLookup: userLookup}
}

func (uc *ListRequisitionsByUserUsecase) Execute(ctx context.Context, query requisitionin.ListRequisitionsByUserQuery) ([]*domain.Requisition, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if query.UserID == "" || !strings.HasPrefix(query.UserID, "usr") {
		return nil, ErrInvalidUserID
	}
	if uc.userLookup != nil {
		exists, err := uc.userLookup.CheckUserExists(ctx, query.UserID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, customerrors.NewNotFoundError("user not found")
		}
	}
	limit, offset := shareddomain.Clamp(query.Limit, query.Offset)
	return uc.requisitionRepo.ListRequisitionsByUserID(ctx, query.UserID, limit, offset)
}

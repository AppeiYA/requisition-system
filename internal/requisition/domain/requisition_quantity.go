package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

type RequisitionQuantity struct {
	quantity int64
}

var (
	ErrInvalidRequisitionQuantity      = customerrors.NewBadRequestError("invalid requisition quantity")
	ErrRequisitionQuantityExceedsLimit = customerrors.NewBadRequestError("requisition quantity exceeds limit")
)

func NewRequisitionQuantity(quantity int) (*RequisitionQuantity, error) {
	if quantity <= 0 {
		return nil, ErrInvalidRequisitionQuantity
	}

	if quantity > 1000 {
		return nil, ErrRequisitionQuantityExceedsLimit
	}

	return &RequisitionQuantity{
		quantity: int64(quantity),
	}, nil
}

func (rq *RequisitionQuantity) Value() int64 {
	return rq.quantity
}

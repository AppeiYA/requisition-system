package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

var (
	ErrInvalidRequisitionJustification = customerrors.NewBadRequestError("invalid requisition justification")
	ErrRequisitionJustificationTooLong = customerrors.NewBadRequestError("requisition justification exceeds maximum length")
)

type RequisitionJustification struct {
	value string
}

func NewRequisitionJustification(value string) (*RequisitionJustification, error) {
	if value == "" {
		return nil, ErrInvalidRequisitionJustification
	}

	if len(value) > 500 {
		return nil, ErrRequisitionJustificationTooLong
	}

	return &RequisitionJustification{
		value: value,
	}, nil
}

func (rj *RequisitionJustification) Value() string {
	return rj.value
}

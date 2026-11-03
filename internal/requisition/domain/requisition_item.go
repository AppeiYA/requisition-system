package domain

import (
	"strings"
	"time"

	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrInvalidRequisitionItemID   = customerrors.NewBadRequestError("invalid requisition item ID")
	ErrInvalidItemSelection       = customerrors.NewBadRequestError("exactly one item or requested item must be selected")
	ErrDuplicateRequisitionItem   = customerrors.NewBadRequestError("catalogue item cannot appear more than once in a requisition")
	ErrDuplicateRequestedItem     = customerrors.NewBadRequestError("requested item cannot belong to more than one line")
	ErrInvalidEstimatedUnitAmount = customerrors.NewBadRequestError("estimated unit amount cannot be negative")
	ErrInvalidLineNumber		  = customerrors.NewBadRequestError("line number must be greater than zero")
)

const RequisitionItemPrefixID = "reqitem"

type RequisitionItem struct {
	id                  string
	requisitionID       string
	lineNumber          int
	itemID              string
	requestedItemID     string
	itemNameSnapshot    *ItemName
	quantity            *RequisitionQuantity
	estimatedUnitAmount *shareddomain.Money
	notes               string
	createdAt           time.Time
}

func NewRequisitionItem(
	id string,
	requisitionID string,
	lineNumber int,
	itemID string,
	requestedItemID string,
	itemNameSnapshot *ItemName,
	quantity *RequisitionQuantity,
	estimatedUnitAmount *shareddomain.Money,
	notes string,
) (*RequisitionItem, error) {
	return newRequisitionItem(id, requisitionID, lineNumber, itemID, requestedItemID,
		itemNameSnapshot, quantity, estimatedUnitAmount, notes, time.Now())
}

func ReconstituteRequisitionItem(
	id string,
	requisitionID string,
	lineNumber int,
	itemID string,
	requestedItemID string,
	itemNameSnapshot *ItemName,
	quantity *RequisitionQuantity,
	estimatedUnitAmount *shareddomain.Money,
	notes string,
	createdAt time.Time,
) (*RequisitionItem, error) {
	return newRequisitionItem(id, requisitionID, lineNumber, itemID, requestedItemID,
		itemNameSnapshot, quantity, estimatedUnitAmount, notes, createdAt)
}

func newRequisitionItem(
	id string,
	requisitionID string,
	lineNumber int,
	itemID string,
	requestedItemID string,
	itemNameSnapshot *ItemName,
	quantity *RequisitionQuantity,
	estimatedUnitAmount *shareddomain.Money,
	notes string,
	createdAt time.Time,
) (*RequisitionItem, error) {
	if id == "" || !strings.HasPrefix(id, RequisitionItemPrefixID) {
		return nil, ErrInvalidRequisitionItemID
	}
	if requisitionID == "" || !strings.HasPrefix(requisitionID, RequisitionPrefixID) {
		return nil, ErrInvalidRequisitionID
	}
	if lineNumber <= 0 {
		return nil, ErrInvalidLineNumber
	}
	if (itemID == "") == (requestedItemID == "") {
		return nil, ErrInvalidItemSelection
	}
	if itemID != "" && !strings.HasPrefix(itemID, ItemPrefixID) {
		return nil, ErrInvalidItemID
	}
	if requestedItemID != "" && !strings.HasPrefix(requestedItemID, RequestedItemPrefixID) {
		return nil, ErrInvalidRequestedItemID
	}
	if itemNameSnapshot == nil {
		return nil, ErrNilFields("itemNameSnapshot")
	}
	if quantity == nil {
		return nil, ErrNilFields("quantity")
	}
	if estimatedUnitAmount == nil {
		return nil, ErrNilFields("estimatedUnitAmount")
	}
	if estimatedUnitAmount.IsNegative() {
		return nil, ErrInvalidEstimatedUnitAmount
	}

	return &RequisitionItem{
		id: id, requisitionID: requisitionID, lineNumber: lineNumber,
		itemID: itemID, requestedItemID: requestedItemID,
		itemNameSnapshot: itemNameSnapshot, quantity: quantity,
		estimatedUnitAmount: estimatedUnitAmount, notes: notes, createdAt: createdAt,
	}, nil
}

func (i *RequisitionItem) ID() string                               { return i.id }
func (i *RequisitionItem) RequisitionID() string                    { return i.requisitionID }
func (i *RequisitionItem) LineNumber() int                          { return i.lineNumber }
func (i *RequisitionItem) ItemID() string                           { return i.itemID }
func (i *RequisitionItem) RequestedItemID() string                  { return i.requestedItemID }
func (i *RequisitionItem) ItemNameSnapshot() *ItemName              { return i.itemNameSnapshot }
func (i *RequisitionItem) Quantity() *RequisitionQuantity           { return i.quantity }
func (i *RequisitionItem) EstimatedUnitAmount() *shareddomain.Money { return i.estimatedUnitAmount }
func (i *RequisitionItem) Notes() string                            { return i.notes }
func (i *RequisitionItem) CreatedAt() time.Time                     { return i.createdAt }

func (i *RequisitionItem) LineTotal() (*shareddomain.Money, error) {
	return i.estimatedUnitAmount.Multiply(i.quantity.Value())
}

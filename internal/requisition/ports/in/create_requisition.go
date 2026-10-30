package requisitionin

import (
	"context"
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type CreateRequisitionCommand struct {
	UserID        string
	ReferenceCode string
	DepartmentID  string
	Lines         []CreateRequisitionLine
	Justification string
	Priority      string
	NeededBy      time.Time
}

// CreateRequisitionLine describes one requisition line. Exactly one of ItemID
// and RequestedItem must be set. EstimatedUnitAmount is expressed in the
// currency's major unit (naira for NGN); persistence converts it to minor
// units.
type CreateRequisitionLine struct {
	ItemID              string
	RequestedItem       *CreateRequestedItem
	Quantity            int
	EstimatedUnitAmount float64
	Notes               string
}

// CreateRequestedItem is the free-filled item payload for a requisition line
// that does not reference the catalogue.
type CreateRequestedItem struct {
	Name        string
	CategoryID  string
	Description string
}

type CreateRequisitionUsecaseInt interface {
	Execute(ctx context.Context, cmd CreateRequisitionCommand) (*domain.Requisition, error)
}

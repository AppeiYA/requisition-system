package requisitionin

import (
	"context"
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type UpdateRequisitionCommand struct {
	RequisitionID string
	UserID        string
	DepartmentID  string
	Lines         []UpdateRequisitionLine
	Justification string
	Priority      string
	NeededBy      time.Time
}

// UpdateRequisitionLine describes one line in a requisition PATCH. When
// RequisitionItemID is set the line must already belong to the requisition
// being updated; otherwise a new line is created.
type UpdateRequisitionLine struct {
	RequisitionItemID   string
	ItemID              string
	RequestedItem       *CreateRequestedItem
	Quantity            int
	EstimatedUnitAmount float64
	Notes               string
}

type UpdateRequisitionUsecaseInt interface {
	Execute(context.Context, UpdateRequisitionCommand) (*domain.Requisition, error)
}

package requisitionin

import (
	"context"
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type CreateRequisitionCommand struct {
	UserID                   string
	ReferenceCode            string
	Lines                    []CreateRequisitionLine
	CategoryID               string
	ItemID                   string
	RequestedItemName        string
	RequestedItemDescription string
	Quantity                 int
	EstimatedCost            float64
	Justification            string
	Priority                 string
	NeededBy                 time.Time
}

type CreateRequisitionLine struct {
	CategoryID               string
	ItemID                   string
	RequestedItemName        string
	RequestedItemDescription string
	Quantity                 int
	EstimatedUnitAmount      float64
	Notes                    string
}

type CreateRequisitionUsecaseInt interface {
	Execute(ctx context.Context, cmd CreateRequisitionCommand) (*domain.Requisition, error)
}

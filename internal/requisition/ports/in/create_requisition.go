package requisitionin

import (
	"context"
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type CreateRequisitionCommand struct {
	UserID                   string
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

type CreateRequisitionUsecaseInt interface {
	Execute(ctx context.Context, cmd CreateRequisitionCommand) (*domain.Requisition, error)
}

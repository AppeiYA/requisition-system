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
	Lines         []CreateRequisitionLine
	Justification string
	Priority      string
	NeededBy      time.Time
}

type UpdateRequisitionUsecaseInt interface {
	Execute(context.Context, UpdateRequisitionCommand) (*domain.Requisition, error)
}

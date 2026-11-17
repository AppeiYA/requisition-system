package itemin

import "context"

type PromoteRequestedItemDTO struct {
	RequestedItemID string
	UnitCost        float64
	ReviewerID      string
}

type PromoteRequestedItemUsecaseInt interface {
	Execute(ctx context.Context, req PromoteRequestedItemDTO) error
}

package itemin

import "context"

type DismissRequestedItemDTO struct {
	RequestedItemID string
	ReviewerID      string
}

type DismissRequestedItemUsecaseInt interface {
	Execute(ctx context.Context, req DismissRequestedItemDTO) error
}

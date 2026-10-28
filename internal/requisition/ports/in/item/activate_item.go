package itemin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type ActivateItemUsecaseInt interface {
	Execute(ctx context.Context, id string) (*domain.Item, error)
}

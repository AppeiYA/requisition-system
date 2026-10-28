package itemin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type DeactivateItemUsecaseInt interface {
	Execute(ctx context.Context, id string) (*domain.Item, error)
}

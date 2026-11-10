package identityin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

type CheckUserStatusUsecaseInt interface {
	Execute(ctx context.Context, userID string) (domain.Status, error)
}
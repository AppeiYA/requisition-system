package adminout

import (
	"context"

	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

type AdminIdentityPort interface {
	CheckUserExists(ctx context.Context, email *shareddomain.Email) (bool, error)
}

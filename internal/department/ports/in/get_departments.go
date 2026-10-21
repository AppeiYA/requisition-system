package departmentin

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/department/domain"
)

type GetDepartmentsUsecaseInt interface {
	Execute(ctx context.Context) ([]*domain.Department, error)
}

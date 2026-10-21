package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/AppeiYA/requisition-system/internal/department/domain"
	departmentout "github.com/AppeiYA/requisition-system/internal/department/ports/out"
)

type GetDepartmentsUsecase struct {
	departmentRepo departmentout.DepartmentRepository
	mu             sync.Mutex
	cached         []*domain.Department
	cacheExpiresAt time.Time
}

func NewGetDepartmentsUsecase(departmentRepo departmentout.DepartmentRepository) *GetDepartmentsUsecase {
	return &GetDepartmentsUsecase{
		departmentRepo: departmentRepo,
	}
}

func (d *GetDepartmentsUsecase) Execute(ctx context.Context) ([]*domain.Department, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	d.mu.Lock()
	if time.Now().Before(d.cacheExpiresAt) {
		result := append([]*domain.Department(nil), d.cached...)
		d.mu.Unlock()
		return result, nil
	}
	d.mu.Unlock()

	departments, err := d.departmentRepo.GetDepartments(ctx)
	if err != nil {
		return nil, err
	}

	d.mu.Lock()
	d.cached = append([]*domain.Department(nil), departments...)
	d.cacheExpiresAt = time.Now().Add(time.Minute)
	d.mu.Unlock()

	return departments, nil
}

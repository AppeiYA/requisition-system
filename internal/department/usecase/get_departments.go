package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/department/domain"
	departmentout "github.com/AppeiYA/requisition-system/internal/department/ports/out"
)

type GetDepartmentsUsecase struct {
	departmentRepo departmentout.DepartmentRepository
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

	departments, err := d.departmentRepo.GetDepartments(ctx)
	if err != nil {
		return nil, err
	}

	return departments, nil
}

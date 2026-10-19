package usecase

import (
	"context"
	"errors"
	"net/http"

	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	departmentout "github.com/AppeiYA/requisition-system/internal/department/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrDepartmentIDRequired = customerrors.NewBadRequestError("department id is required")
	ErrDepartmentNotFound   = customerrors.NewNotFoundError("department not found")
)

type GetDepartmentByIDUsecase struct {
	departmentRepo departmentout.DepartmentRepository
}

func NewGetDepartmentByIDUsecase(departmentRepo departmentout.DepartmentRepository) *GetDepartmentByIDUsecase {
	return &GetDepartmentByIDUsecase{departmentRepo: departmentRepo}
}

func (uc *GetDepartmentByIDUsecase) Execute(ctx context.Context, id string) (*departmentin.DepartmentResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if id == "" {
		return nil, ErrDepartmentIDRequired
	}

	department, err := uc.departmentRepo.GetDepartmentByID(ctx, id)
	if err != nil {
		var custErr *customerrors.CustomError
		if errors.As(err, &custErr) && custErr.Code == http.StatusNotFound {
			return nil, ErrDepartmentNotFound
		}
		return nil, err
	}

	return &departmentin.DepartmentResult{
		ID:          department.ID(),
		Name:        department.Name().Value(),
		Description: department.Description().Value(),
		Code:        department.Code().Value(),
		Status:      department.Status().String(),
		CreatedAt:   department.CreatedAt(),
	}, nil
}
